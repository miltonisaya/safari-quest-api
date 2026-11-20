package fileresource

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"safari-quest-api/config"
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
//	@Summary		List file resources
//	@Description	Files the caller may attach to a package. Administrators see every company.
//	@Tags			File Resources
//	@Produce		json
//	@Success		200	{object}	response.CustomApiResponse{data=List}
//	@Security		BearerAuth
//	@Router			/file-resources [get]
func (controller Controller) Index(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to fetch file resources")
		return
	}
	items, err := GetAll(pagination.FromContext(c), scope)
	if err != nil {
		fail(c, err, "Failed to fetch file resources")
		return
	}
	response.Success(c, http.StatusOK, "File resources retrieved", items)
}

// Show godoc
//
//	@Summary	Get file resource
//	@Tags		File Resources
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource UUID"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Security	BearerAuth
//	@Router		/file-resources/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to fetch file resource")
	if !ok {
		return
	}
	item, err := GetByUUID(uid, scope)
	if err != nil {
		fail(c, err, "Failed to fetch file resource")
		return
	}
	response.Success(c, http.StatusOK, "File resource retrieved", item)
}

// Create godoc
//
//	@Summary		Upload a file
//	@Description	Multipart fields: file, name, code, file_resource_type_uuid, operator_uuid
//	@Tags			File Resources
//	@Accept			mpfd
//	@Produce		json
//	@Success		201	{object}	response.CustomApiResponse{data=Response}
//	@Failure		400	{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/file-resources [post]
func (controller Controller) Create(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to upload file")
		return
	}
	var form struct {
		Name     string `form:"name" binding:"required"`
		Code     string `form:"code" binding:"required"`
		TypeID   string `form:"file_resource_type_uuid" binding:"required,uuid"`
		Operator string `form:"operator_uuid" binding:"required,uuid"`
	}
	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	input := Input{
		Name:                 form.Name,
		Code:                 form.Code,
		FileResourceTypeUUID: uuid.MustParse(form.TypeID),
		OperatorUUID:         uuid.MustParse(form.Operator),
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Choose a file to upload", nil)
		return
	}
	path, err := Store(header, config.App.UploadDir)
	if err != nil {
		fail(c, err, "Failed to upload file")
		return
	}
	item, err := Create(input, path, scope)
	if err != nil {
		_ = os.Remove(filepath.Join(config.App.UploadDir, filepath.Base(path)))
		fail(c, err, "Failed to upload file")
		return
	}
	response.Success(c, http.StatusCreated, "File uploaded", item)
}

// Update godoc
//
//	@Summary	Update file resource details
//	@Tags		File Resources
//	@Accept		json
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource UUID"
//	@Param		body	body		Input	true	"File resource"
//	@Success	200		{object}	response.CustomApiResponse{data=Response}
//	@Security	BearerAuth
//	@Router		/file-resources/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to update file resource")
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
		fail(c, err, "Failed to update file resource")
		return
	}
	response.Success(c, http.StatusOK, "File resource updated", item)
}

// Delete godoc
//
//	@Summary	Delete file resource
//	@Tags		File Resources
//	@Produce	json
//	@Param		uuid	path		string	true	"File resource UUID"
//	@Success	200		{object}	response.CustomApiResponse
//	@Security	BearerAuth
//	@Router		/file-resources/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to delete file resource")
	if !ok {
		return
	}
	if err := Delete(uid, scope); err != nil {
		fail(c, err, "Failed to delete file resource")
		return
	}
	response.Success(c, http.StatusOK, "File resource deleted", nil)
}

// Serve sends an uploaded file. The name is the file on disk, not the resource UUID.
func Serve(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	if name == "." || name == string(filepath.Separator) || strings.Contains(name, "..") {
		c.Status(http.StatusNotFound)
		return
	}
	path := filepath.Join(config.App.UploadDir, name)
	if _, err := os.Stat(path); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.File(path)
}

func scopeAndID(c *gin.Context, fallback string) (companyscope.Scope, uuid.UUID, bool) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, fallback)
		return companyscope.Scope{}, uuid.Nil, false
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid file resource UUID", nil)
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
		response.Fail(c, http.StatusNotFound, "File resource not found", nil)
	case errors.Is(err, ErrOutsideCompany), errors.Is(err, companyscope.ErrNoCompany):
		response.Fail(c, http.StatusForbidden, ErrOutsideCompany.Error(), nil)
	case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrFileType), errors.Is(err, ErrFileTooLarge):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, errUnauthenticated):
		response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
	default:
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
