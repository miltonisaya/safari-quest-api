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

type MenuController struct{}

// Index godoc
//
//	@Summary		List menus
//	@Description	Return a paginated, searchable, and sortable list of menus
//	@Tags			Menus
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name"
//	@Param			sort_by		query		string	false	"Sort column: name, sort_order, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"							default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[services.MenuResponse]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/menus [get]
func (controller MenuController) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	menus, err := services.MenuGetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch menus")
		return
	}
	response.Success(c, http.StatusOK, "Menus retrieved", menus)
}

// Show godoc
//
//	@Summary		Get menu
//	@Description	Return a single menu by UUID
//	@Tags			Menus
//	@Produce		json
//	@Param			uuid	path		string	true	"Menu UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=services.MenuResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/menus/{uuid} [get]
func (controller MenuController) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid menu UUID", nil)
		return
	}
	menu, err := services.MenuGetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Menu not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch menu")
		return
	}
	response.Success(c, http.StatusOK, "Menu retrieved", menu)
}

// Create godoc
//
//	@Summary		Create menu
//	@Description	Create a new menu
//	@Tags			Menus
//	@Accept			json
//	@Produce		json
//	@Param			body	body		services.MenuInput	true	"Menu data"
//	@Success		201		{object}	response.CustomApiResponse{data=services.MenuResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/menus [post]
func (controller MenuController) Create(c *gin.Context) {
	var input services.MenuInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	menu, err := services.MenuCreate(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create menu")
		return
	}
	response.Success(c, http.StatusCreated, "Menu created", menu)
}

// Update godoc
//
//	@Summary		Update menu
//	@Description	Update an existing menu by UUID
//	@Tags			Menus
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Menu UUID"
//	@Param			body	body		services.MenuInput	true	"Menu data"
//	@Success		200		{object}	response.CustomApiResponse{data=services.MenuResponse}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/menus/{uuid} [put]
func (controller MenuController) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid menu UUID", nil)
		return
	}
	var input services.MenuInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	menu, err := services.MenuUpdate(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Menu not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update menu")
		return
	}
	response.Success(c, http.StatusOK, "Menu updated", menu)
}

// Delete godoc
//
//	@Summary		Delete menu
//	@Description	Soft-delete a menu by UUID
//	@Tags			Menus
//	@Produce		json
//	@Param			uuid	path		string	true	"Menu UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/menus/{uuid} [delete]
func (controller MenuController) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid menu UUID", nil)
		return
	}
	if err := services.MenuDelete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "Menu not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete menu")
		return
	}
	response.Success(c, http.StatusOK, "Menu deleted", nil)
}
