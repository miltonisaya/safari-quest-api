package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/pkg/mailer"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// Queues. Failed sends wait in a retry queue (its TTL) and then return to
// QueueEmails; after MaxAttempts a message goes to QueueDead for inspection.
const (
	QueueEmails = "emails"
	QueueDead   = "emails.dead"
	MaxAttempts = 5
)

// retryDelays[i] is the wait after the (i+1)th failed attempt.
var retryDelays = []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 30 * time.Minute}

func retryQueue(d time.Duration) string { return fmt.Sprintf("emails.retry.%ds", int(d.Seconds())) }

// job is the message body: which outbox row to send.
type job struct {
	ID string `json:"id"`
}

// declare creates the queues (idempotent).
func declare(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(QueueDead, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(QueueEmails, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": QueueDead,
	}); err != nil {
		return err
	}
	for _, d := range retryDelays {
		if _, err := ch.QueueDeclare(retryQueue(d), true, false, false, false, amqp.Table{
			"x-message-ttl":             int64(d / time.Millisecond),
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueEmails,
		}); err != nil {
			return err
		}
	}
	return nil
}

func publish(ctx context.Context, ch *amqp.Channel, queue, id string) error {
	body, _ := json.Marshal(job{ID: id})
	confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    id,
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		return err
	}
	ok, err := confirm.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("broker did not confirm the message")
	}
	return nil
}

// withChannel connects, declares the queues and runs fn until the connection
// drops, then reconnects after a pause, until ctx ends.
func withChannel(ctx context.Context, name string, fn func(ctx context.Context, ch *amqp.Channel) error) {
	for ctx.Err() == nil {
		err := func() error {
			conn, err := amqp.Dial(config.App.AMQPURL)
			if err != nil {
				return err
			}
			defer conn.Close()
			ch, err := conn.Channel()
			if err != nil {
				return err
			}
			defer ch.Close()
			if err := declare(ch); err != nil {
				return err
			}
			log.Printf("[email] %s connected to RabbitMQ\n", name)
			return fn(ctx, ch)
		}()
		if ctx.Err() != nil {
			return
		}
		log.Printf("[email] %s: %v; retrying in 5s\n", name, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// ── Relay: outbox → RabbitMQ ────────────────────────────────────────────────

// RelayEvery is how often the relay looks for unpublished emails.
const RelayEvery = 2 * time.Second

// RunRelay publishes pending outbox rows to RabbitMQ. While RabbitMQ is
// unreachable the rows simply wait, so no email is lost.
func RunRelay(ctx context.Context) {
	withChannel(ctx, "relay", func(ctx context.Context, ch *amqp.Channel) error {
		if err := ch.Confirm(false); err != nil {
			return err
		}
		closed := ch.NotifyClose(make(chan *amqp.Error, 1))
		ticker := time.NewTicker(RelayEvery)
		defer ticker.Stop()
		for {
			if err := relayBatch(ctx, ch); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return nil
			case err := <-closed:
				return fmt.Errorf("channel closed: %v", err)
			case <-ticker.C:
			}
		}
	})
}

// relayBatch publishes one batch; each row is marked queued only after
// RabbitMQ confirms it.
func relayBatch(ctx context.Context, ch *amqp.Channel) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		rows, err := claimPending(tx, 50)
		if err != nil {
			return err
		}
		for _, r := range rows {
			if err := publish(ctx, ch, QueueEmails, r.UUID.String()); err != nil {
				return err
			}
			if err := markQueued(tx, r.ID, time.Now()); err != nil {
				return err
			}
		}
		return nil
	})
}

// ── Worker: RabbitMQ → SMTP ─────────────────────────────────────────────────

// RunWorker sends queued emails until ctx ends.
func RunWorker(ctx context.Context) {
	withChannel(ctx, "worker", func(ctx context.Context, ch *amqp.Channel) error {
		if err := ch.Confirm(false); err != nil {
			return err
		}
		if err := ch.Qos(10, 0, false); err != nil {
			return err
		}
		deliveries, err := ch.Consume(QueueEmails, "", false, false, false, false, nil)
		if err != nil {
			return err
		}
		for {
			select {
			case <-ctx.Done():
				return nil
			case d, ok := <-deliveries:
				if !ok {
					return errors.New("delivery channel closed")
				}
				handle(ctx, ch, d)
			}
		}
	})
}

// send delivers one outbox row. It is a variable so tests can replace it.
var send = func(row Outbox) error {
	var c Content
	if err := json.Unmarshal([]byte(row.Content), &c); err != nil {
		return err
	}
	html, err := c.HTML()
	if err != nil {
		return err
	}
	return mailer.SendMessage(mailer.Message{To: row.ToAddress, Subject: c.Subject, Text: c.Text(), HTML: html})
}

func handle(ctx context.Context, ch *amqp.Channel, d amqp.Delivery) {
	var j job
	if err := json.Unmarshal(d.Body, &j); err != nil {
		_ = d.Nack(false, false) // unreadable: straight to the dead-letter queue
		return
	}
	row, err := findByUUID(j.ID)
	if err != nil || row.Status == StatusSent || row.Status == StatusFailed {
		_ = d.Ack(false) // unknown or already handled (redelivery): nothing to do
		return
	}
	if err := send(row); err == nil {
		if err := markSent(row.ID, time.Now()); err != nil {
			log.Printf("[email] sent %s but could not record it: %v\n", row.UUID, err)
		}
		_ = d.Ack(false)
		return
	} else {
		attempts := row.Attempts + 1
		final := attempts >= MaxAttempts
		reason := err.Error()
		if len(reason) > 500 {
			reason = reason[:500]
		}
		if dbErr := markAttemptFailed(row.ID, attempts, reason, final); dbErr != nil {
			_ = d.Nack(false, true) // try again later rather than lose track
			return
		}
		if final {
			log.Printf("[email] giving up on %s (%s) after %d attempts: %s\n", row.UUID, row.Kind, attempts, reason)
			_ = d.Nack(false, false) // dead-letter queue
			return
		}
		delay := retryDelays[min(attempts-1, len(retryDelays)-1)]
		if err := publish(ctx, ch, retryQueue(delay), row.UUID.String()); err != nil {
			_ = d.Nack(false, true)
			return
		}
		log.Printf("[email] %s (%s) failed (attempt %d), retrying in %s: %s\n", row.UUID, row.Kind, attempts, delay, reason)
		_ = d.Ack(false)
	}
}
