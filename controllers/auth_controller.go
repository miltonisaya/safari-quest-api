package controllers

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/response"
	"safari-quest-api/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (controller AuthController) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}

	result, err := services.Login(input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			response.Fail(c, http.StatusUnauthorized, err.Error(), nil)
		case errors.Is(err, services.ErrAccountInactive):
			response.Fail(c, http.StatusForbidden, err.Error(), nil)
		default:
			response.Error(c, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}
