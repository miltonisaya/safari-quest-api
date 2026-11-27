package account

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/internal/auth"
	"safari-quest-api/internal/email"
	"safari-quest-api/internal/user"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Password reset by a one-time code sent by email.
const (
	CodeTTL      = 10 * time.Minute
	CodeAttempts = 5           // wrong guesses before the code stops working
	resendGap    = time.Minute // stops the form being used to flood an inbox
)

var ErrInvalidCode = errors.New("That code is incorrect or has expired. Request a new code if you need one.")

// PasswordResetCode is a one-time reset code. Only an HMAC of the code is
// stored, keyed with the app secret, so a database leak can't reveal codes.
type PasswordResetCode struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	CodeHash  string
	Attempts  int `gorm:"not null;default:0"`
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type ForgotInput struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type ResetInput struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Code     string `json:"code" binding:"required,max=12"`
	Password string `json:"password" binding:"required,max=72"` // bcrypt reads at most 72 bytes
}

// newCode returns a uniformly random 6-digit code.
func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashCode binds the code to the user, so the same digits differ per account.
func hashCode(userID uint, code string) string {
	m := hmac.New(sha256.New, []byte(config.App.JWTSecret))
	fmt.Fprintf(m, "%d:%s", userID, code)
	return hex.EncodeToString(m.Sum(nil))
}

func codeEmail(u user.User, code string) email.Draft {
	return email.Draft{Kind: "password_reset_code", To: u.Email, Content: email.Content{
		Subject:    "Your SafariQuest password reset code: " + code,
		Preheader:  "Use this code within 10 minutes to choose a new password.",
		Heading:    "Reset your password",
		Paragraphs: []string{fmt.Sprintf("Hello %s, use this code to choose a new SafariQuest password:", u.FirstName)},
		Code:       code,
		CodeNote:   fmt.Sprintf("It works once, for %d minutes.", int(CodeTTL.Minutes())),
		After:      []string{"If you didn't ask to reset your password, ignore this email. Your password stays the same, and nobody can change it without this code."},
	}}
}

func changedEmail(u user.User) email.Draft {
	return email.Draft{Kind: "password_changed", To: u.Email, Content: email.Content{
		Subject: "Your SafariQuest password was changed",
		Heading: "Your password was changed",
		Paragraphs: []string{
			fmt.Sprintf("Hello %s, the password for your SafariQuest account was just reset with a code sent to this address.", u.FirstName),
			"If this wasn't you, reset your password again straight away and contact us.",
		},
		Button: &email.Button{Label: "Sign in", URL: strings.TrimRight(config.App.AppURL, "/") + "/login"},
	}}
}

// ForgotPassword emails a 6-digit code to an active account. The caller gets
// the same answer whether or not the email has an account, so the form can't
// be used to find out who is registered; the email itself goes out
// asynchronously through the queue.
func ForgotPassword(input ForgotInput) error {
	account, err := user.FindByEmail(input.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !account.IsActive) {
		return nil
	}
	if err != nil {
		return err
	}
	code, err := newCode()
	if err != nil {
		return err
	}
	now := time.Now()
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var recent int64
		if err := tx.Model(&PasswordResetCode{}).
			Where("user_id = ? AND created_at > ?", account.ID, now.Add(-resendGap)).Count(&recent).Error; err != nil {
			return err
		}
		if recent > 0 {
			return nil
		}
		// A new code cancels any earlier one.
		if err := tx.Where("user_id = ? AND used_at IS NULL", account.ID).Delete(&PasswordResetCode{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&PasswordResetCode{
			UserID: account.ID, CodeHash: hashCode(account.ID, code), ExpiresAt: now.Add(CodeTTL),
		}).Error; err != nil {
			return err
		}
		return email.Enqueue(tx, codeEmail(account, code))
	})
}

// ResetPassword sets a new password when the code matches. Wrong guesses are
// counted; after CodeAttempts the code stops working.
func ResetPassword(input ResetInput) error {
	if !auth.StrongPassword(input.Password) {
		return ErrWeakPassword
	}
	code := strings.TrimSpace(input.Code)
	account, err := user.FindByEmail(input.Email)
	if err != nil || !account.IsActive || len(code) != 6 {
		return ErrInvalidCode
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now()
	wrong := false
	err = database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var c PasswordResetCode
		// Lock the code so two guesses can't race past the attempt limit.
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND used_at IS NULL AND expires_at > ? AND attempts < ?", account.ID, now, CodeAttempts).
			Order("created_at DESC").Take(&c).Error
		if err != nil {
			return ErrInvalidCode
		}
		if !hmac.Equal([]byte(c.CodeHash), []byte(hashCode(account.ID, code))) {
			wrong = true
			return tx.Model(&c).Update("attempts", gorm.Expr("attempts + 1")).Error
		}
		if err := tx.Model(&user.User{}).Where("id = ?", account.ID).Update("password", string(hashed)).Error; err != nil {
			return err
		}
		if err := tx.Model(&c).Update("used_at", now).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND used_at IS NULL", account.ID).Delete(&PasswordResetCode{}).Error; err != nil {
			return err
		}
		return email.Enqueue(tx, changedEmail(account))
	})
	if wrong {
		return ErrInvalidCode
	}
	return err
}
