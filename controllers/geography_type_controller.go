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

type GeographyTypeController struct{}

// Index godoc
//
//	@Summary		List geography types
//	@Description	Return a paginated, searchable, and sortable list of geography types
//	@Tags			Geography Types
//	@Produce		json
//	@Param			page		query		int		false	"Page number"			default(1)
//	@Param			per_page	query		int		false	"Items per page"		default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[services.GeographyTypeResponse]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geography-types [get]
func (controller GeographyTypeController) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	g, err := services.GeographyTypeGetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geography types")
		return
	}
	response.Success(c, http.StatusOK, "Geography types retrieved", g)
}

// Show godoc
//
//	@Summary		Get geography type
//	@Description	Return a single geography type by UUID
//	@Tags			Geography Types
//	@Produce		json
//	@Param			uuid	path		string	true	"Geography type UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=services.GeographyTypeResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geography-types/{uuid} [get]
func (controller GeographyTypeController) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography type UUID", nil)
		return
	}
	g, err := services.GeographyTypeGetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geography type")
		return
	}
	response.Success(c, http.StatusOK, "Geography type retrieved", g)
}

// Create godoc
//
//	@Summary		Create geography type
//	@Description	Create a new geography type
//	@Tags			Geography Types
//	@Accept			json
//	@Produce		json
//	@Param			body	body		services.GeographyTypeInput		true	"Geography type data"
//	@Success		201		{object}	response.CustomApiResponse{data=services.GeographyTypeResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geography-types [post]
func (controller GeographyTypeController) Create(c *gin.Context) {
	var input services.GeographyTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	geographyType, err := services.GeographyTypeCreate(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create geography type")
		return
	}
	response.Success(c, http.StatusCreated, "Geography type created", geographyType)
}

// Update godoc
//
//	@Summary		Update geography type
//	@Description	Update an existing geography type by UUID
//	@Tags			Geography Types
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string							true	"Geography type UUID"
//	@Param			body	body		services.GeographyTypeInput		true	"Geography type data"
//	@Success		200		{object}	response.CustomApiResponse{data=services.GeographyTypeResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geography-types/{uuid} [put]
func (controller GeographyTypeController) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography type UUID", nil)
		return
	}
	var input services.GeographyTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	geographyType, err := services.GeographyTypeUpdate(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update geography type")
		return
	}
	response.Success(c, http.StatusOK, "Geography type updated", geographyType)
}

// Delete godoc
//
//	@Summary		Delete geography type
//	@Description	Soft-delete a geography type by UUID
//	@Tags			Geography Types
//	@Produce		json
//	@Param			uuid	path		string	true	"Geography type UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geography-types/{uuid} [delete]
func (controller GeographyTypeController) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography type UUID", nil)
		return
	}
	if err := services.GeographyTypeDelete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete geography type")
		return
	}
	response.Success(c, http.StatusOK, "Geography type deleted", nil)
}
