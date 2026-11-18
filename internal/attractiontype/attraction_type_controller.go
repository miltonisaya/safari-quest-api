package attractionType

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
//	@Summary		List attraction types
//	@Description	Return a paginated, searchable, and sortable list of attraction types
//	@Tags			Attraction Types
//	@Produce		json
//	@Param			page		query		int		false	"Page number"			default(1)
//	@Param			per_page	query		int		false	"Items per page"		default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-types [get]
func (controller Controller) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	g, err := GetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch attraction types")
		return
	}
	response.Success(c, http.StatusOK, "Attraction types retrieved", g)
}

// Show godoc
//
//	@Summary		Get attraction type
//	@Description	Return a single attraction type by UUID
//	@Tags			Attraction Types
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction type UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-types/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction type UUID", nil)
		return
	}
	g, err := GetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch attraction type")
		return
	}
	response.Success(c, http.StatusOK, "Attraction type retrieved", g)
}

// Create godoc
//
//	@Summary		Create attraction type
//	@Description	Create a new attraction type
//	@Tags			Attraction Types
//	@Accept			json
//	@Produce		json
//	@Param			body	body		Input		true	"Attraction type data"
//	@Success		201		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-types [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	attraction, err := Create(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create attraction type")
		return
	}
	response.Success(c, http.StatusCreated, "Attraction type created", attraction)
}

// Update godoc
//
//	@Summary		Update attraction type
//	@Description	Update an existing attraction type by UUID
//	@Tags			Attraction Types
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string							true	"Attraction type UUID"
//	@Param			body	body		Input		true	"Attraction type data"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-types/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction type UUID", nil)
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	attraction, err := Update(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update attraction type")
		return
	}
	response.Success(c, http.StatusOK, "Attraction type updated", attraction)
}

// Delete godoc
//
//	@Summary		Delete attraction type
//	@Description	Soft-delete a attraction type by UUID
//	@Tags			Attraction Types
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction type UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-types/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction type UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction type not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete attraction type")
		return
	}
	response.Success(c, http.StatusOK, "Attraction type deleted", nil)
}
