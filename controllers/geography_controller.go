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

type GeographyController struct{}

// Index godoc
//
//	@Summary		List geographies
//	@Description	Return a paginated list of geographies. Pass parent_uuid to fetch children of a specific geography; omit it to fetch top-level geographies.
//	@Tags			Geographies
//	@Produce		json
//	@Param			parent_uuid	query		string	false	"Parent geography UUID — omit to return top-level geographies"
//	@Param			page		query		int		false	"Page number"			default(1)
//	@Param			per_page	query		int		false	"Items per page"		default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[services.GeographyResponse]}
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geographies [get]
func (controller GeographyController) Index(c *gin.Context) {
	params := pagination.FromContext(c)

	var parentUUID *uuid.UUID
	if raw := c.Query("parent_uuid"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "Invalid parent_uuid", nil)
			return
		}
		parentUUID = &parsed
	}

	g, err := services.GeographyGetAll(params, parentUUID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geographies")
		return
	}
	response.Success(c, http.StatusOK, "Geographies retrieved", g)
}

// Show godoc
//
//	@Summary		Get geography
//	@Description	Return a single geography by UUID, including its direct children
//	@Tags			Geographies
//	@Produce		json
//	@Param			uuid	path		string	true	"Geography UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=services.GeographyResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geographies/{uuid} [get]
func (controller GeographyController) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography UUID", nil)
		return
	}
	g, err := services.GeographyGetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geography")
		return
	}
	response.Success(c, http.StatusOK, "Geography retrieved", g)
}

// Create godoc
//
//	@Summary		Create geography
//	@Description	Create a new geography. Provide parentUuid to nest it under an existing geography.
//	@Tags			Geographies
//	@Accept			json
//	@Produce		json
//	@Param			body	body		services.GeographyInput		true	"Geography data"
//	@Success		201		{object}	response.CustomApiResponse{data=services.GeographyResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geographies [post]
func (controller GeographyController) Create(c *gin.Context) {
	var input services.GeographyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	geography, err := services.GeographyCreate(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create geography")
		return
	}
	response.Success(c, http.StatusCreated, "Geography created", geography)
}

// Update godoc
//
//	@Summary		Update geography
//	@Description	Update an existing geography by UUID
//	@Tags			Geographies
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string					true	"Geography UUID"
//	@Param			body	body		services.GeographyInput	true	"Geography data"
//	@Success		200		{object}	response.CustomApiResponse{data=services.GeographyResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geographies/{uuid} [put]
func (controller GeographyController) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography UUID", nil)
		return
	}
	var input services.GeographyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	geography, err := services.GeographyUpdate(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update geography")
		return
	}
	response.Success(c, http.StatusOK, "Geography updated", geography)
}

// Delete godoc
//
//	@Summary		Delete geography
//	@Description	Soft-delete a geography by UUID
//	@Tags			Geographies
//	@Produce		json
//	@Param			uuid	path		string	true	"Geography UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/geographies/{uuid} [delete]
func (controller GeographyController) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid geography UUID", nil)
		return
	}
	if err := services.GeographyDelete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Geography not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete geography")
		return
	}
	response.Success(c, http.StatusOK, "Geography deleted", nil)
}
