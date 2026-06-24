package services

import (
	"errors"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/repositories"

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
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// Login validates the provided credentials and, on success, returns a signed JWT
// together with the user profile wrapped inside LoginResponse.
// The caller (controller) passes LoginResponse.Data into response.Success so the
// final JSON is enclosed in CustomApiResponse as with every other endpoint.
func Login(input LoginInput) (LoginResponse, error) {
	user, err := repositories.UserFindByEmail(input.Email)
	if err != nil {
		// Treat record-not-found the same as wrong password to avoid leaking
		// whether an email address is registered in the system.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginResponse{}, ErrInvalidCredentials
		}
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	if !user.IsActive {
		return LoginResponse{}, ErrAccountInactive
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.UUID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: tokenStr,
		User:  toUserResponse(user),
	}, nil
}
