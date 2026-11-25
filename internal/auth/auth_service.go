package auth

import (
	"errors"
	"safari-quest-api/internal/authority"
	"safari-quest-api/internal/user"
	"time"

	"safari-quest-api/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Sentinel errors let the controller map failures to the correct HTTP status
// without inspecting error strings.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is inactive")
)

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token       string                   `json:"token"`
	User        user.Response            `json:"user"`
	Authorities []authority.ListResponse `json:"authorities"`
}

// Login validates the provided credentials and, on success, returns a signed JWT
// together with the user profile wrapped inside LoginResponse.
// The caller (controller) passes LoginResponse.Data into response.Success so the
// final JSON is enclosed in CustomApiResponse as with every other endpoint.
func Login(input LoginInput) (LoginResponse, error) {
	account, err := user.FindByEmail(input.Email)
	if err != nil {
		// Treat record-not-found the same as wrong password to avoid leaking
		// whether an email address is registered in the system.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	if !account.IsActive {
		return LoginResponse{}, ErrAccountInactive
	}

	return issueSession(account)
}

// issueSession signs a 24-hour JWT for the account and lists the authority
// codes its roles grant. Roles must be loaded with their authorities.
func issueSession(account user.User) (LoginResponse, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": account.UUID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		return LoginResponse{}, err
	}

	seen := make(map[string]bool)
	authorities := make([]authority.ListResponse, 0)
	for _, role := range account.Roles {
		for _, a := range role.Authorities {
			if !seen[a.Code] {
				seen[a.Code] = true
				authorities = append(authorities, authority.ListResponse{
					UUID:      a.UUID,
					Name:      a.Name,
					Code:      a.Code,
					CreatedAt: a.CreatedAt,
					UpdatedAt: a.UpdatedAt,
				})
			}
		}
	}

	return LoginResponse{
		Token:       tokenStr,
		User:        user.ToResponse(account),
		Authorities: authorities,
	}, nil
}
