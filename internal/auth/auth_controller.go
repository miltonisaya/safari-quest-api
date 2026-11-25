package auth

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticate with email and password and receive a JWT token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginInput	true	"Login credentials"
//	@Success		200		{object}	response.CustomApiResponse{data=LoginResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		401		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Router			/auth/login [post]
func (controller Controller) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}

	result, err := Login(input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.Fail(c, http.StatusUnauthorized, err.Error(), nil)
		case errors.Is(err, ErrAccountInactive):
			response.Fail(c, http.StatusForbidden, err.Error(), nil)
		default:
			response.Error(c, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}

// Register godoc
//
//	@Summary		Create an account
//	@Description	Sign up as a traveler, or as a tour operator with a new company (unverified until an administrator verifies it; its packages stay off the public site until then). Signs the new account in.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		auth.RegisterInput	true	"Account details"
//	@Success		201		{object}	response.CustomApiResponse{data=auth.RegisterResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		503		{object}	response.CustomApiResponse
//	@Router			/auth/register [post]
func (controller Controller) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Please check the highlighted fields", gin.H{"error": err.Error()})
		return
	}
	result, err := Register(input)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken), errors.Is(err, ErrCompanyTaken):
			response.Fail(c, http.StatusConflict, err.Error(), nil)
		case errors.Is(err, ErrSignupDisabled):
			response.Fail(c, http.StatusServiceUnavailable, err.Error(), nil)
		case errors.Is(err, ErrWeakPassword), errors.Is(err, ErrTermsRequired), errors.Is(err, ErrCompanyRequired),
			errors.Is(err, ErrInvalidName), errors.Is(err, ErrInvalidAccountTy):
			response.Fail(c, http.StatusBadRequest, err.Error(), nil)
		default:
			response.Error(c, http.StatusInternalServerError, "Could not create your account")
		}
		return
	}
	message := "Welcome to SafariQuest"
	if result.CompanyPendingReview {
		message = "Account created. Your company will appear on SafariQuest once it is verified."
	}
	response.Success(c, http.StatusCreated, message, result)
}
