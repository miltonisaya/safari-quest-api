package attraction

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
//	@Summary		List attractions
//	@Description	Return a paginated, searchable, sortable and filterable list of attractions
//	@Tags			Attractions
//	@Produce		json
//	@Param			page						query		int		false	"Page number"		default(1)
//	@Param			per_page					query		int		false	"Items per page"	default(15)
//	@Param			search						query		string	false	"Search by name or code"
//	@Param			sort_by						query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir					query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Param			attraction_type_uuid		query		string	false	"Only attractions of this type"
//	@Param			attraction_category_uuid	query		string	false	"Only attractions in this category"
//	@Param			geography_uuid				query		string	false	"Only attractions in this geography"
//	@Param			tour_operator_uuid			query		string	false	"Only attractions offered by this tour operator"
//	@Success		200							{object}	response.CustomApiResponse{data=pagination.Result[attraction.Response]}
//	@Failure		400							{object}	response.CustomApiResponse
//	@Failure		500							{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attractions [get]
func (controller Controller) Index(c *gin.Context) {
	var f Filters
	for _, q := range []struct {
		name string
		dst  **uuid.UUID
	}{
		{"attraction_type_uuid", &f.AttractionTypeUUID},
		{"attraction_category_uuid", &f.AttractionCategoryUUID},
		{"geography_uuid", &f.GeographyUUID},
		{"tour_operator_uuid", &f.TourOperatorUUID},
	} {
		if v := c.Query(q.name); v != "" {
			id, err := uuid.Parse(v)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, "Invalid "+q.name, nil)
				return
			}
			*q.dst = &id
		}
	}
	attractions, err := GetAll(pagination.FromContext(c), f)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch attractions")
		return
	}
	response.Success(c, http.StatusOK, "Attractions retrieved", attractions)
}

// Show godoc
//
//	@Summary		Get attraction
//	@Description	Return a single attraction by UUID, with its type, category, geography and tour operators
//	@Tags			Attractions
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=attraction.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attractions/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction UUID", nil)
		return
	}
	a, err := GetByUUID(uid)
	if err != nil {
		respondError(c, err, "Failed to fetch attraction")
		return
	}
	response.Success(c, http.StatusOK, "Attraction retrieved", a)
}

// Create godoc
//
//	@Summary		Create attraction
//	@Description	Create a new attraction linked to a type, category, geography and optional tour operators
//	@Tags			Attractions
//	@Accept			json
//	@Produce		json
//	@Param			body	body		attraction.Input	true	"Attraction data"
//	@Success		201		{object}	response.CustomApiResponse{data=attraction.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attractions [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	a, err := Create(input)
	if err != nil {
		respondError(c, err, "Failed to create attraction")
		return
	}
	response.Success(c, http.StatusCreated, "Attraction created", a)
}

// Update godoc
//
//	@Summary		Update attraction
//	@Description	Update an attraction by UUID; tour_operator_uuids replaces the linked operators
//	@Tags			Attractions
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Attraction UUID"
//	@Param			body	body		attraction.Input	true	"Attraction data"
//	@Success		200		{object}	response.CustomApiResponse{data=attraction.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attractions/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction UUID", nil)
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	a, err := Update(uid, input)
	if err != nil {
		respondError(c, err, "Failed to update attraction")
		return
	}
	response.Success(c, http.StatusOK, "Attraction updated", a)
}

// Delete godoc
//
//	@Summary		Delete attraction
//	@Description	Soft-delete an attraction by UUID
//	@Tags			Attractions
//	@Produce		json
//	@Param			uuid	path		string	true	"Attraction UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/attractions/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid attraction UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		respondError(c, err, "Failed to delete attraction")
		return
	}
	response.Success(c, http.StatusOK, "Attraction deleted", nil)
}

func respondError(c *gin.Context, err error, failed string) {
	switch {
	case errors.Is(err, ErrInvalidReference):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "Attraction not found", nil)
	default:
		response.Error(c, http.StatusInternalServerError, failed)
	}
}
