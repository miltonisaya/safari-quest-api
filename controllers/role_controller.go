package controllers

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/pagination"
	"safari-quest-api/pkg/response"
	"safari-quest-api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleController struct{}

// Index godoc
// @Summary      List roles
// @Description  Return a paginated, searchable, and sortable list of roles
// @Tags         Roles
// @Produce      json
// @Param        page     query int    false "Page number"          default(1)
// @Param        per_page query int    false "Items per page"       default(15)
// @Param        search   query string false "Search by name or code"
// @Param        sort_by  query string false "Sort column: name, code, created_at, updated_at" default(created_at)
// @Param        sort_dir query string false "Sort direction: asc or desc"                      default(asc)
// @Success      200 {object} response.CustomApiResponse{data=pagination.Result[services.RoleResponse]}
// @Failure      500 {object} response.CustomApiResponse
// @Security     BearerAuth
// @Router       /roles [get]
func (controller RoleController) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	roles, err := services.RoleGetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	response.Success(c, http.StatusOK, "Roles retrieved", roles)
}

// Show godoc
// @Summary      Get role
// @Description  Return a single role by UUID
// @Tags         Roles
// @Produce      json
// @Param        uuid path string true "Role UUID"
// @Success      200 {object} response.CustomApiResponse{data=services.RoleResponse}
// @Failure      400 {object} response.CustomApiResponse
// @Failure      404 {object} response.CustomApiResponse
// @Failure      500 {object} response.CustomApiResponse
// @Security     BearerAuth
// @Router       /roles/{uuid} [get]
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

// Create godoc
// @Summary      Create role
// @Description  Create a new role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        body body services.RoleInput true "Role data"
// @Success      201 {object} response.CustomApiResponse{data=services.RoleResponse}
// @Failure      400 {object} response.CustomApiResponse
// @Failure      500 {object} response.CustomApiResponse
// @Security     BearerAuth
// @Router       /roles [post]
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

// Update godoc
// @Summary      Update role
// @Description  Update an existing role by UUID
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Param        uuid path string true "Role UUID"
// @Param        body body services.RoleInput true "Role data"
// @Success      200 {object} response.CustomApiResponse{data=services.RoleResponse}
// @Failure      400 {object} response.CustomApiResponse
// @Failure      404 {object} response.CustomApiResponse
// @Failure      500 {object} response.CustomApiResponse
// @Security     BearerAuth
// @Router       /roles/{uuid} [put]
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

// Delete godoc
// @Summary      Delete role
// @Description  Soft-delete a role by UUID
// @Tags         Roles
// @Produce      json
// @Param        uuid path string true "Role UUID"
// @Success      200 {object} response.CustomApiResponse
// @Failure      400 {object} response.CustomApiResponse
// @Failure      404 {object} response.CustomApiResponse
// @Failure      500 {object} response.CustomApiResponse
// @Security     BearerAuth
// @Router       /roles/{uuid} [delete]
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
