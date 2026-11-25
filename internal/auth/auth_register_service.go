package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Account types a person can sign up as.
const (
	AccountTraveler = "traveler"
	AccountOperator = "operator"
)

var (
	ErrEmailTaken       = errors.New("An account with this email already exists. Sign in instead.")
	ErrCompanyTaken     = errors.New("A company with this name is already on SafariQuest. Ask its staff to add you, or use your registered trading name.")
	ErrWeakPassword     = errors.New("Use at least 8 characters, including a letter and a number")
	ErrTermsRequired    = errors.New("Please accept the Terms of Service and Privacy Policy")
	ErrCompanyRequired  = errors.New("Enter your company name and phone number")
	ErrSignupDisabled   = errors.New("Sign-up for this account type isn't available yet")
	ErrInvalidName      = errors.New("Enter your full name")
	ErrInvalidAccountTy = errors.New("Choose traveler or operator")
)

// RegisterInput is what the "Create your account" form sends.
type RegisterInput struct {
	FullName    string `json:"full_name" binding:"required,max=200"`
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,max=72"` // bcrypt reads at most 72 bytes
	AccountType string `json:"account_type" binding:"required"`
	AcceptTerms bool   `json:"accept_terms"`
	// Operators only:
	CompanyName string `json:"company_name" binding:"max=255"`
	Phone       string `json:"phone" binding:"max=30"`
}

// RegisterResponse signs the new account in, like login, and says whether
// the company still waits for verification.
type RegisterResponse struct {
	LoginResponse
	AccountType          string `json:"account_type"`
	CompanyPendingReview bool   `json:"company_pending_review"`
}

// strongPassword: 8+ characters with at least one letter and one digit.
func strongPassword(pw string) bool {
	if len(pw) < 8 {
		return false
	}
	var letter, digit bool
	for _, r := range pw {
		switch {
		case unicode.IsLetter(r):
			letter = true
		case unicode.IsDigit(r):
			digit = true
		}
	}
	return letter && digit
}

// splitName turns "Juma Ali Mwangi" into ("Juma Ali", "Mwangi").
func splitName(full string) (string, string) {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return strings.Join(parts[:len(parts)-1], " "), parts[len(parts)-1]
	}
}

var nonCode = regexp.MustCompile(`[^A-Z0-9]+`)

// companyCode makes a unique-enough code like "SERENGETI_TRAILS_3F9A".
func companyCode(name string) string {
	base := strings.Trim(nonCode.ReplaceAllString(strings.ToUpper(name), "_"), "_")
	if len(base) > 40 {
		base = strings.Trim(base[:40], "_")
	}
	if base == "" {
		base = "COMPANY"
	}
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return base + "_" + strings.ToUpper(hex.EncodeToString(b))
}

// Register creates an account and signs it in. Travelers get the Traveler
// role. Operators also get a new company (unverified, so its packages stay off
// the public site until an administrator verifies it) and the Tour Operator role.
func Register(input RegisterInput) (RegisterResponse, error) {
	kind := strings.ToLower(strings.TrimSpace(input.AccountType))
	if kind != AccountTraveler && kind != AccountOperator {
		return RegisterResponse{}, ErrInvalidAccountTy
	}
	if !input.AcceptTerms {
		return RegisterResponse{}, ErrTermsRequired
	}
	first, last := splitName(input.FullName)
	if first == "" {
		return RegisterResponse{}, ErrInvalidName
	}
	if !strongPassword(input.Password) {
		return RegisterResponse{}, ErrWeakPassword
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	company, phone := strings.TrimSpace(input.CompanyName), strings.TrimSpace(input.Phone)
	if kind == AccountOperator && (company == "" || phone == "") {
		return RegisterResponse{}, ErrCompanyRequired
	}

	taken, err := emailTaken(email)
	if err != nil {
		return RegisterResponse{}, err
	}
	if taken {
		return RegisterResponse{}, ErrEmailTaken
	}
	if kind == AccountOperator {
		exists, err := companyNameTaken(company)
		if err != nil {
			return RegisterResponse{}, err
		}
		if exists {
			return RegisterResponse{}, ErrCompanyTaken
		}
	}

	roleCode := "TRAVELER"
	if kind == AccountOperator {
		roleCode = touroperator.OperatorRoleCode
	}
	roleRow, err := roleByCode(roleCode)
	if err != nil {
		return RegisterResponse{}, ErrSignupDisabled
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, err
	}
	account := user.User{
		FirstName: first,
		LastName:  last,
		Email:     email,
		Password:  string(hashed),
		Mobile:    phone,
		IsActive:  true,
	}
	var operator *touroperator.TourOperator
	if kind == AccountOperator {
		operator = &touroperator.TourOperator{
			Name:       company,
			Code:       companyCode(company),
			Email:      email,
			Phone:      phone,
			IsActive:   true,
			IsVerified: false,
		}
	}
	// A unique violation is either the email (another sign-up won the race)
	// or, rarely, the random company code; retry the latter with a new code.
	for attempt := 0; ; attempt++ {
		err = createAccount(&account, roleRow, operator)
		if err == nil {
			break
		}
		if !isUniqueViolation(err) {
			return RegisterResponse{}, err
		}
		if taken, _ := emailTaken(email); taken || operator == nil || attempt >= 3 {
			return RegisterResponse{}, ErrEmailTaken
		}
		if exists, _ := companyNameTaken(company); exists {
			return RegisterResponse{}, ErrCompanyTaken
		}
		account.ID, account.UUID = 0, uuid.Nil
		operator.ID, operator.UUID = 0, uuid.Nil
		operator.Code = companyCode(company)
	}

	// Reload with roles and authorities, exactly as login does.
	saved, err := user.FindByEmail(email)
	if err != nil {
		return RegisterResponse{}, err
	}
	session, err := issueSession(saved)
	if err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{
		LoginResponse:        session,
		AccountType:          kind,
		CompanyPendingReview: operator != nil,
	}, nil
}
