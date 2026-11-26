package account

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/internal/auth"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/mailer"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// ResetLinkTTL is how long an emailed reset link works.
	ResetLinkTTL = time.Hour
	// resendGap stops the form from being used to flood someone's inbox.
	resendGap = time.Minute
)

var (
	ErrWeakPassword    = auth.ErrWeakPassword
	ErrInvalidLink     = errors.New("This reset link is invalid or has expired. Request a new one.")
	ErrWrongPassword   = errors.New("Your current password is incorrect")
	ErrSamePassword    = errors.New("Choose a password different from your current one")
	ErrInvalidName     = errors.New("Enter your first and last name")
	ErrInvalidSex      = errors.New("Choose Male or Female")
	ErrAccountNotFound = errors.New("account not found")
)

// ── Password reset ──────────────────────────────────────────────────────────

type ForgotInput struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type ResetInput struct {
	Token    string `json:"token" binding:"required,max=200"`
	Password string `json:"password" binding:"required,max=72"` // bcrypt reads at most 72 bytes
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ForgotPassword emails a reset link if the email belongs to an active
// account. The caller always gets the same answer, so the form cannot be used
// to find out who has an account; the email goes out in the background for
// the same reason (response time).
func ForgotPassword(input ForgotInput) error {
	account, err := user.FindByEmail(input.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !account.IsActive) {
		return nil
	}
	if err != nil {
		return err
	}
	if recent, err := issuedSince(account.ID, time.Now().Add(-resendGap)); err != nil || recent {
		return err
	}
	token, err := newToken()
	if err != nil {
		return err
	}
	if err := insertToken(&PasswordResetToken{
		UserID:    account.ID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(ResetLinkTTL),
	}); err != nil {
		return err
	}
	link := strings.TrimRight(config.App.AppURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
	go func() {
		if err := mailer.Send(account.Email, "Reset your SafariQuest password", resetEmail(account.FirstName, link)); err != nil {
			log.Printf("[account] could not send reset email: %v\n", err)
		}
	}()
	return nil
}

func resetEmail(firstName, link string) string {
	return fmt.Sprintf(`Hello %s,

We received a request to reset your SafariQuest password. Open this link within %d minutes to choose a new one:

%s

If you didn't ask for this, you can ignore this email; your password stays the same.

The SafariQuest team
`, firstName, int(ResetLinkTTL.Minutes()), link)
}

// ResetPassword sets a new password from an emailed link. A link works once.
func ResetPassword(input ResetInput) error {
	if !auth.StrongPassword(input.Password) {
		return ErrWeakPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = redeem(hashToken(strings.TrimSpace(input.Token)), string(hashed), time.Now())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrInvalidLink
	}
	return err
}

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
