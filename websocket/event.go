package websocket

import "encoding/json"

type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

// Event is the JSON envelope pushed to every connected client whenever a
// mutation occurs. Entity names the resource ("menu", "role", "user") and
// Action describes what happened. Payload carries the response DTO so the
// client can update its local state without issuing a follow-up GET.
type Event struct {
	Entity  string `json:"entity"`
	Action  Action `json:"action"`
	Payload any    `json:"payload"`
}

func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
