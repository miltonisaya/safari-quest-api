package catalog

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"safari-quest-api/internal/review"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Controller struct{}

// Index godoc
//
//	@Summary		Public package catalog
//	@Description	Upcoming packages for the website, with rating, photo and seats left. No sign-in required.
//	@Tags			Catalog
//	@Produce		json
//	@Param			category	query		string	false	"Package type, e.g. Safari, Beach, Cultural"
//	@Param			eco			query		bool	false	"Only eco-certified packages"
//	@Param			max_price	query		number	false	"Highest price per person"
//	@Param			search		query		string	false	"Name, description, location or type"
//	@Param			sort		query		string	false	"soonest (default), price-asc, price-desc or newest"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page (max 100)"	default(100)
//	@Success		200			{object}	response.CustomApiResponse{data=catalog.List}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Router			/catalog/packages [get]
func (controller Controller) Index(c *gin.Context) {
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(MaxPerPage)))
	list, err := Packages(tourpackage.CatalogFilters{
		Category:     strings.TrimSpace(c.Query("category")),
		EcoCertified: c.Query("eco") == "true",
		MaxPrice:     maxPrice,
		Search:       strings.TrimSpace(c.Query("search")),
		Sort:         c.Query("sort"),
	}, page, perPage)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch packages")
		return
	}
	response.Success(c, http.StatusOK, "Packages retrieved", list)
}

// Show godoc
//
//	@Summary		Public package detail
//	@Description	Everything the website shows about one package. No sign-in required.
//	@Tags			Catalog
//	@Produce		json
//	@Param			uuid	path		string	true	"Package UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=catalog.Package}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Router			/catalog/packages/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	item, err := PackageByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Package not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch the package")
		return
	}
	response.Success(c, http.StatusOK, "Package retrieved", item)
}

// Reviews godoc
//
//	@Summary		Public package reviews
//	@Description	Visible reviews of a package with its rating summary. No sign-in required.
//	@Tags			Catalog
//	@Produce		json
//	@Param			uuid		path		string	true	"Package UUID"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Success		200			{object}	response.CustomApiResponse{data=review.PackageList}
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		404			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Router			/catalog/packages/{uuid}/reviews [get]
func (controller Controller) Reviews(c *gin.Context) {
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	list, err := Reviews(uid, pagination.FromContext(c))
	if err != nil {
		if errors.Is(err, review.ErrInvalidReference) {
			response.Fail(c, http.StatusNotFound, "Package not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch reviews")
		return
	}
	response.Success(c, http.StatusOK, "Reviews retrieved", list)
}

func pathUUID(c *gin.Context) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid package UUID", nil)
		return uuid.Nil, false
	}
	return uid, true
}
