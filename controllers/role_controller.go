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

func (rc RoleController) Index(c *gin.Context) {
	roles, err := services.RoleGetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	response.Success(c, http.StatusOK, "Roles retrieved", roles)
}

func (rc RoleController) Show(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role ID", nil)
		return
	}
	role, err := services.RoleGetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch role")
		return
	}
	response.Success(c, http.StatusOK, "Role retrieved", role)
}

func (rc RoleController) Create(c *gin.Context) {
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

func (rc RoleController) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role ID", nil)
		return
	}
	var input services.RoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	role, err := services.RoleUpdate(id, input)
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

func (rc RoleController) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role ID", nil)
		return
	}
	if err := services.RoleDelete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	response.Success(c, http.StatusOK, "Role deleted", nil)
}
