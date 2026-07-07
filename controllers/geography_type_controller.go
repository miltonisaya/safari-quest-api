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

func (controller GeographyTypeController) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	g, err := services.GeographyTypeGetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geography type types")
		return
	}
	response.Success(c, http.StatusOK, "Geographies retrieved", g)
}

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
