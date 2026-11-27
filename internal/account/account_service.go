package account

import (
	"errors"
	"strings"

	"safari-quest-api/internal/auth"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrWeakPassword    = auth.ErrWeakPassword
	ErrWrongPassword   = errors.New("Your current password is incorrect")
	ErrSamePassword    = errors.New("Choose a password different from your current one")
	ErrInvalidName     = errors.New("Enter your first and last name")
	ErrInvalidSex      = errors.New("Choose Male or Female")
	ErrAccountNotFound = errors.New("account not found")
)

// ── Own profile ─────────────────────────────────────────────────────────────

// ProfileInput is what a user may change about themselves. Email, roles and
// active status stay with administrators.
type ProfileInput struct {
	FirstName  string `json:"first_name" binding:"required,max=100"`
	MiddleName string `json:"middle_name" binding:"max=100"`
	LastName   string `json:"last_name" binding:"required,max=100"`
	Sex        string `json:"sex" binding:"max=10"` // Male or Female, checked in UpdateProfile for a clear message
	Mobile     string `json:"mobile" binding:"max=30"`
	Address    string `json:"address" binding:"max=255"`
}

type PasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required,max=72"`
	NewPassword     string `json:"new_password" binding:"required,max=72"`
}

func load(id uuid.UUID) (user.User, error) {
	u, err := user.FindByUUID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return u, ErrAccountNotFound
	}
	return u, err
}

func Profile(id uuid.UUID) (user.Response, error) {
	u, err := load(id)
	if err != nil {
		return user.Response{}, err
	}
	return user.ToResponse(u), nil
}

func UpdateProfile(id uuid.UUID, input ProfileInput) (user.Response, error) {
	first, last := strings.TrimSpace(input.FirstName), strings.TrimSpace(input.LastName)
	if first == "" || last == "" {
		return user.Response{}, ErrInvalidName
	}
	sex := strings.TrimSpace(input.Sex)
	if sex != "Male" && sex != "Female" {
		return user.Response{}, ErrInvalidSex
	}
	u, err := load(id)
	if err != nil {
		return user.Response{}, err
	}
	if err := saveProfile(u.ID, map[string]any{
		"first_name":  first,
		"middle_name": strings.TrimSpace(input.MiddleName),
		"last_name":   last,
		"sex":         sex,
		"mobile":      strings.TrimSpace(input.Mobile),
		"address":     strings.TrimSpace(input.Address),
	}); err != nil {
		return user.Response{}, err
	}
	return Profile(id)
}

// ChangePassword needs the current password, so a session left open on a
// shared computer can't be used to take over the account.
// shortcut: tokens already issued stay valid until they expire (24h); add a token version to sign out other devices.
func ChangePassword(id uuid.UUID, input PasswordInput) error {
	u, err := load(id)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.CurrentPassword)) != nil {
		return ErrWrongPassword
	}
	if input.NewPassword == input.CurrentPassword {
		return ErrSamePassword
	}
	if !auth.StrongPassword(input.NewPassword) {
		return ErrWeakPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return savePassword(u.ID, string(hashed))
}
