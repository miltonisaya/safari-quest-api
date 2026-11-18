package user

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
//	@Summary		List users
//	@Description	Return a paginated, searchable, and sortable list of users
//	@Tags			Users
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by first name, last name, email, or mobile"
//	@Param			sort_by		query		string	false	"Sort column: first_name, last_name, email, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"										default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users [get]
func (controller Controller) Index(c *gin.Context) {
	params := pagination.FromContext(c)
	users, err := GetAll(params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	response.Success(c, http.StatusOK, "Users retrieved", users)
}

// Show godoc
//
//	@Summary		Get user
//	@Description	Return a single user by UUID
//	@Tags			Users
//	@Produce		json
//	@Param			uuid	path		string	true	"User UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	user, err := GetByUUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch user")
		return
	}
	response.Success(c, http.StatusOK, "User retrieved", user)
}

// Create godoc
//
//	@Summary		Create user
//	@Description	Create a new user and assign roles
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		Input	true	"User data"
//	@Success		201		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	user, err := Create(input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}
	response.Success(c, http.StatusCreated, "User created", user)
}

// Update godoc
//
//	@Summary		Update user
//	@Description	Update an existing user by UUID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"User UUID"
//	@Param			body	body		Input	true	"User data"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	user, err := Update(uid, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update user")
		return
	}
	response.Success(c, http.StatusOK, "User updated", user)
}

// Delete godoc
//
//	@Summary		Delete user
//	@Description	Soft-delete a user by UUID
//	@Tags			Users
//	@Produce		json
//	@Param			uuid	path		string	true	"User UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	if err := Delete(uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "User not found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}
	response.Success(c, http.StatusOK, "User deleted", nil)
}
