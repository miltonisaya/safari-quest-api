package controllers

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/response"
	"safari-quest-api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserController struct{}

func (controller UserController) Index(c *gin.Context) {
	users, err := services.UserGetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	response.Success(c, http.StatusOK, "Users retrieved", users)
}

func (controller UserController) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	user, err := services.UserGetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}
	response.Success(c, http.StatusOK, "User retrieved", user)
}

func (controller UserController) Create(c *gin.Context) {
	var input services.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	user, err := services.UserCreate(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}
	response.Success(c, http.StatusCreated, "User created", user)
}

func (controller UserController) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	var input services.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	user, err := services.UserUpdate(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}
	response.Success(c, http.StatusOK, "User updated", user)
}

func (controller UserController) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	if err := services.UserDelete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}
	response.Success(c, http.StatusOK, "User deleted", nil)
}
