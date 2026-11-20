package fileresourcetype

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
//	@Summary		List file resource types
//	@Description	Return a paginated list of file resource types
//	@Tags			File Resource Types
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name or code"
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/file-resource-types [get]
func (controller Controller) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	items, err := GetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch file resource types")
		return
	}
	response.Success(c, http.StatusOK, "File resource types retrieved", items)
}

// Show godoc
//
//	@Summary	Get file resource type
//	@Tags		File Resource Types
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource type UUID"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Failure	400		{object}	response.CustomApiResponse
//	@Failure	404		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/file-resource-types/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid file resource type UUID", nil)
		return
	}
	item, err := GetByUUID(uid)
	if err != nil {
		fail(c, err, "Failed to fetch file resource type", "File resource type not found")
		return
	}
	response.Success(c, http.StatusOK, "File resource type retrieved", item)
}

// Create godoc
//
//	@Summary	Create file resource type
//	@Tags		File Resource Types
//	@Accept		json
//	@Produce	json
//	@Param		body	body		Input	true	"File resource type"
//	@Success	201		{object}	response.CustomApiResponse{data=Response}
//	@Failure	400		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/file-resource-types [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Create(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create file resource type")
		return
	}
	response.Success(c, http.StatusCreated, "File resource type created", item)
}

// Update godoc
//
//	@Summary	Update file resource type
//	@Tags		File Resource Types
//	@Accept		json
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource type UUID"
//	@Param		body	body		Input	true	"File resource type"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Failure	400		{object}	response.CustomApiResponse
//	@Failure	404		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/file-resource-types/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid file resource type UUID", nil)
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Update(uid, input)
	if err != nil {
		fail(c, err, "Failed to update file resource type", "File resource type not found")
		return
	}
	response.Success(c, http.StatusOK, "File resource type updated", item)
}

// Delete godoc
//
//	@Summary	Delete file resource type
//	@Tags		File Resource Types
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource type UUID"
//	@Success	200		{object}	response.CustomApiResponse
//	@Failure	404		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/file-resource-types/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid file resource type UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		fail(c, err, "Failed to delete file resource type", "File resource type not found")
		return
	}
	response.Success(c, http.StatusOK, "File resource type deleted", nil)
}

func fail(c *gin.Context, err error, fallback, missing string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, http.StatusNotFound, missing, nil)
		return
	}
	response.Error(c, http.StatusInternalServerError, fallback)
}
