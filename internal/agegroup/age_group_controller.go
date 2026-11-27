package agegroup

import (
	"errors"
	"net/http"

	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List age groups
//	@Description	The age bands travellers are put in, youngest first. Public: the booking form needs them.
//	@Tags			Age groups
//	@Produce		json
//	@Success		200	{object}	response.CustomApiResponse{data=[]agegroup.Response}
//	@Router			/catalog/age-groups [get]
func (Controller) Index(c *gin.Context) {
	items, err := GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Could not load age groups")
		return
	}
	response.Success(c, http.StatusOK, "Age groups", items)
}

// Replace godoc
//
//	@Summary		Replace the age groups
//	@Description	Saves the whole list at once. Groups must start at 0 and run without gaps or overlaps, with only the oldest open-ended. Groups used by upcoming bookings can't be removed.
//	@Tags			Age groups
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		agegroup.ReplaceInput	true	"All age groups"
//	@Success		200		{object}	response.CustomApiResponse{data=[]agegroup.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Router			/age-groups [put]
func (Controller) Replace(c *gin.Context) {
	var input ReplaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, ErrInvalidGroups.Error(), gin.H{"error": err.Error()})
		return
	}
	items, err := ReplaceAll(input)
	switch {
	case errors.Is(err, ErrInvalidGroups):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, ErrGroupInUse):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case err != nil:
		response.Error(c, http.StatusInternalServerError, "Could not save age groups")
	default:
		response.Success(c, http.StatusOK, "Age groups saved", items)
	}
}
