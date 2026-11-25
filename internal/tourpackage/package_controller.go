package tourpackage

import (
	"errors"
	"net/http"

	"safari-quest-api/internal/companyscope"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List packages
//	@Description	Trips for sale. A tour company sees only its own packages.
//	@Tags			Packages
//	@Produce		json
//	@Success		200	{object}	response.CustomApiResponse{data=List}
//	@Security		BearerAuth
//	@Router			/packages [get]
func (controller Controller) Index(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to fetch packages")
		return
	}
	items, err := GetAll(pagination.FromContext(c), scope)
	if err != nil {
		fail(c, err, "Failed to fetch packages")
		return
	}
	response.Success(c, http.StatusOK, "Packages retrieved", items)
}

// Show godoc
//
//	@Summary	Get package
//	@Tags		Packages
//	@Produce	json
//	@Param		uuid	path		string	true	"Package UUID"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Security	BearerAuth
//	@Router		/packages/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to fetch package")
	if !ok {
		return
	}
	item, err := GetByUUID(uid, scope)
	if err != nil {
		fail(c, err, "Failed to fetch package")
		return
	}
	response.Success(c, http.StatusOK, "Package retrieved", item)
}

// Create godoc
//
//	@Summary	Create package
//	@Tags		Packages
//	@Accept		json
//	@Produce	json
//	@Param		body	body		Input	true	"Package"
//	@Success	201		{object}	response.CustomApiResponse{data=Response}
//	@Security	BearerAuth
//	@Router		/packages [post]
func (controller Controller) Create(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to create package")
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Create(input, scope)
	if err != nil {
		fail(c, err, "Failed to create package")
		return
	}
	response.Success(c, http.StatusCreated, "Package created", item)
}

// Update godoc
//
//	@Summary	Update package
//	@Tags		Packages
//	@Accept		json
//	@Produce	json
//	@Param		uuid	path		string	true	"Package UUID"
//	@Param		body	body		Input	true	"Package"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Security	BearerAuth
//	@Router		/packages/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to update package")
	if !ok {
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Update(uid, input, scope)
	if err != nil {
		fail(c, err, "Failed to update package")
		return
	}
	response.Success(c, http.StatusOK, "Package updated", item)
}

// Delete godoc
//
//	@Summary	Delete package
//	@Tags		Packages
//	@Produce	json
//	@Param		uuid	path		string	true	"Package UUID"
//	@Success	200		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/packages/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to delete package")
	if !ok {
		return
	}
	if err := Delete(uid, scope); err != nil {
		fail(c, err, "Failed to delete package")
		return
	}
	response.Success(c, http.StatusOK, "Package deleted", nil)
}

func scopeAndID(c *gin.Context, fallback string) (companyscope.Scope, uuid.UUID, bool) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, fallback)
		return companyscope.Scope{}, uuid.Nil, false
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid package UUID", nil)
		return companyscope.Scope{}, uuid.Nil, false
	}
	return scope, uid, true
}

func loadScope(c *gin.Context) (companyscope.Scope, error) {
	value, ok := c.Get("userUUID")
	if !ok {
		return companyscope.Scope{}, errUnauthenticated
	}
	id, ok := value.(uuid.UUID)
	if !ok {
		return companyscope.Scope{}, errUnauthenticated
	}
	return companyscope.Load(id)
}

var errUnauthenticated = errors.New("Unauthorized")

func fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "Package not found", nil)
	case errors.Is(err, ErrOutsideCompany), errors.Is(err, companyscope.ErrNoCompany):
		response.Fail(c, http.StatusForbidden, ErrOutsideCompany.Error(), nil)
	case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrInvalidPackage), errors.Is(err, ErrAttachment),
		errors.Is(err, ErrInvalidItinerary), errors.Is(err, ErrInvalidDetails):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, errUnauthenticated):
		response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
	default:
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
