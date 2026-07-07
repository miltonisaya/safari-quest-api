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

func (controller GeographyController) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	g, err := services.GeographyGetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch geographies")
		return
	}
	response.Success(c, http.StatusOK, "Geographies retrieved", g)
}

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
