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

type RoleController struct{}

func (controller RoleController) Index(c *gin.Context) {
	roles, err := services.RoleGetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	response.Success(c, http.StatusOK, "Roles retrieved", roles)
}

func (controller RoleController) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid authority UUID", nil)
		return
	}
	authority, err := services.RoleGetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch role")
		return
	}
	response.Success(c, http.StatusOK, "Role retrieved", authority)
}

func (controller RoleController) Create(c *gin.Context) {
	var input services.RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	role, err := services.RoleCreate(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}
	response.Success(c, http.StatusCreated, "Role created", role)
}

func (controller RoleController) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	var input services.RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	role, err := services.RoleUpdate(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update role")
		return
	}
	response.Success(c, http.StatusOK, "Role updated", role)
}

func (controller RoleController) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	if err := services.RoleDelete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Authority not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	response.Success(c, http.StatusOK, "Authority deleted", nil)
}
