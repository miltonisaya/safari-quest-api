package account

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct{}

// ForgotPassword godoc
//
//	@Summary		Request a password reset code
//	@Description	Emails a 6-digit code (valid 10 minutes) through the email queue if the address belongs to an active account. Always answers the same way, so it does not reveal who has an account.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		account.ForgotInput	true	"Account email"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Router			/auth/forgot-password [post]
func (Controller) ForgotPassword(c *gin.Context) {
	var input ForgotInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Enter a valid email address", gin.H{"error": err.Error()})
		return
	}
	if err := ForgotPassword(input); err != nil {
		response.Error(c, http.StatusInternalServerError, "Could not send the reset link. Please try again.")
		return
	}
	response.Success(c, http.StatusOK, "If an account uses that email, we've sent it a 6-digit code.", nil)
}

// ResetPassword godoc
//
//	@Summary		Set a new password with an emailed code
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		account.ResetInput	true	"Email, the 6-digit code and the new password"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Router			/auth/reset-password [post]
func (Controller) ResetPassword(c *gin.Context) {
	var input ResetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Enter your email, the 6-digit code and a new password", gin.H{"error": err.Error()})
		return
	}
	if err := ResetPassword(input); err != nil {
		fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Your password has been reset. You can sign in now.", nil)
}

// Me godoc
//
//	@Summary		Get my profile
//	@Tags			Account
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.CustomApiResponse{data=user.Response}
//	@Failure		401	{object}	response.CustomApiResponse
//	@Router			/me [get]
func (Controller) Me(c *gin.Context) {
	profile, err := Profile(requester(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profile", profile)
}

// UpdateMe godoc
//
//	@Summary		Update my profile
//	@Description	Changes your own name, sex, mobile and address. Email, roles and status are managed by administrators.
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		account.ProfileInput	true	"Profile"
//	@Success		200		{object}	response.CustomApiResponse{data=user.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		401		{object}	response.CustomApiResponse
//	@Router			/me [put]
func (Controller) UpdateMe(c *gin.Context) {
	var input ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Please check the highlighted fields", gin.H{"error": err.Error()})
		return
	}
	profile, err := UpdateProfile(requester(c), input)
	if err != nil {
		fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Profile updated", profile)
}

// ChangeMyPassword godoc
//
//	@Summary		Change my password
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		account.PasswordInput	true	"Current and new password"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		401		{object}	response.CustomApiResponse
//	@Router			/me/password [put]
func (Controller) ChangeMyPassword(c *gin.Context) {
	var input PasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Enter your current and new password", gin.H{"error": err.Error()})
		return
	}
	if err := ChangePassword(requester(c), input); err != nil {
		fail(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Password changed", nil)
}

func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrWeakPassword), errors.Is(err, ErrInvalidCode), errors.Is(err, ErrWrongPassword),
		errors.Is(err, ErrSamePassword), errors.Is(err, ErrInvalidName), errors.Is(err, ErrInvalidSex):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, ErrAccountNotFound):
		response.Fail(c, http.StatusNotFound, "Account not found", nil)
	default:
		response.Error(c, http.StatusInternalServerError, "Something went wrong. Please try again.")
	}
}

func requester(c *gin.Context) uuid.UUID {
	if v, ok := c.Get("userUUID"); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}
