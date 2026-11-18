package role

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/pagination"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List roles
//	@Description	Return a paginated, searchable, and sortable list of roles
//	@Tags			Roles
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles [get]
func (controller Controller) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	roles, err := GetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}
	response.Success(c, http.StatusOK, "Roles retrieved", roles)
}

// Show godoc
//
//	@Summary		Get role
//	@Description	Return a single role by UUID
//	@Tags			Roles
//	@Produce		json
//	@Param			uuid	path		string	true	"Role UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid authority UUID", nil)
		return
	}
	authority, err := GetByUUID(uid)
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
//
//	@Summary		Create role
//	@Description	Create a new role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			body	body		Input	true	"Role data"
//	@Success		201		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	role, err := Create(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create role")
		return
	}
	response.Success(c, http.StatusCreated, "Role created", role)
}

// Update godoc
//
//	@Summary		Update role
//	@Description	Update an existing role by UUID
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Role UUID"
//	@Param			body	body		Input	true	"Role data"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	role, err := Update(uid, input)
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

// UpdateAuthorities godoc
//
//	@Summary		Update role authorities
//	@Description	Replace the full set of authorities assigned to a role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string							true	"Role UUID"
//	@Param			body	body		AuthoritiesInput	true	"Authority UUIDs"
//	@Success		200		{object}	response.CustomApiResponse{data=[]authority.ListResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles/{uuid}/authorities [put]
func (controller Controller) UpdateAuthorities(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	var input AuthoritiesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	authorities, err := UpdateAuthorities(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update role authorities")
		return
	}
	response.Success(c, http.StatusOK, "Role authorities updated", authorities)
}

// Authorities godoc
//
//	@Summary		List role authorities
//	@Description	Return all authorities assigned to a role
//	@Tags			Roles
//	@Produce		json
//	@Param			uuid	path		string	true	"Role UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=[]authority.ListResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles/{uuid}/authorities [get]
func (controller Controller) Authorities(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	authorities, err := GetAuthorities(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Role not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch role authorities")
		return
	}
	response.Success(c, http.StatusOK, "Role authorities retrieved", authorities)
}

// Delete godoc
//
//	@Summary		Delete role
//	@Description	Soft-delete a role by UUID
//	@Tags			Roles
//	@Produce		json
//	@Param			uuid	path		string	true	"Role UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/roles/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid role UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Authority not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete role")
		return
	}
	response.Success(c, http.StatusOK, "Authority deleted", nil)
}
