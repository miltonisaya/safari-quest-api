package attractioncategory

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
//	@Summary		List attraction categories
//	@Description	Return a paginated, searchable, and sortable list of attraction categories
//	@Tags			Attraction Categories
//	@Produce		json
//	@Param			page		query		int		false	"Page number"			default(1)
//	@Param			per_page	query		int		false	"Items per page"		default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-categories [get]
func (controller Controller) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	categories, err := GetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch attraction categories")
		return
	}
	response.Success(c, http.StatusOK, "Attraction categories retrieved", categories)
}

// Show godoc
//
//	@Summary		Get attraction category
//	@Description	Return a single attraction category by UUID
//	@Tags			Attraction Categories
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction category UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-categories/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction category UUID", nil)
		return
	}
	category, err := GetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction category not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch attraction category")
		return
	}
	response.Success(c, http.StatusOK, "Attraction category retrieved", category)
}

// Create godoc
//
//	@Summary		Create attraction category
//	@Description	Create a new attraction category
//	@Tags			Attraction Categories
//	@Accept			json
//	@Produce		json
//	@Param			body	body		Input		true	"Attraction category data"
//	@Success		201		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-categories [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	category, err := Create(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create attraction category")
		return
	}
	response.Success(c, http.StatusCreated, "Attraction category created", category)
}

// Update godoc
//
//	@Summary		Update attraction category
//	@Description	Update an existing attraction category by UUID
//	@Tags			Attraction Categories
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string							true	"Attraction category UUID"
//	@Param			body	body		Input		true	"Attraction category data"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-categories/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction category UUID", nil)
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	category, err := Update(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction category not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update attraction category")
		return
	}
	response.Success(c, http.StatusOK, "Attraction category updated", category)
}

// Delete godoc
//
//	@Summary		Delete attraction category
//	@Description	Soft-delete an attraction category by UUID
//	@Tags			Attraction Categories
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction category UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attraction-categories/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction category UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Attraction category not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete attraction category")
		return
	}
	response.Success(c, http.StatusOK, "Attraction category deleted", nil)
}
