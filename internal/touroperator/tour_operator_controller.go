package touroperator

import (
	"errors"
	"net/http"

	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Controller struct{}

// Index godoc
//
//	@Summary		List tour operators
//	@Description	Return a paginated, searchable, and sortable list of tour operators
//	@Tags			Tour Operators
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name, code, legal name or email"
//	@Param			sort_by		query		string	false	"Sort column: name, code, created_at, updated_at"	default(created_at)
//	@Param			sort_dir	query		string	false	"Sort direction: asc or desc"						default(asc)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[touroperator.Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators [get]
func (controller Controller) Index(c *gin.Context) {
	operators, err := GetAll(pagination.FromContext(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch tour operators")
		return
	}
	response.Success(c, http.StatusOK, "Tour operators retrieved", operators)
}

// Show godoc
//
//	@Summary		Get tour operator
//	@Description	Return a single tour operator by UUID
//	@Tags			Tour Operators
//	@Produce		json
//	@Param			uuid	path		string	true	"Tour operator UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=touroperator.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	operator, err := GetByUUID(uid)
	if err != nil {
		respondError(c, err, "Tour operator not found", "Failed to fetch tour operator")
		return
	}
	response.Success(c, http.StatusOK, "Tour operator retrieved", operator)
}

// Create godoc
//
//	@Summary		Create tour operator
//	@Description	Register a new tour operator company
//	@Tags			Tour Operators
//	@Accept			json
//	@Produce		json
//	@Param			body	body		touroperator.Input	true	"Tour operator data"
//	@Success		201		{object}	response.CustomApiResponse{data=touroperator.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	operator, err := Create(input)
	if err != nil {
		respondError(c, err, "Geography not found", "Failed to create tour operator")
		return
	}
	response.Success(c, http.StatusCreated, "Tour operator created", operator)
}

// Update godoc
//
//	@Summary		Update tour operator
//	@Description	Update an existing tour operator by UUID
//	@Tags			Tour Operators
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Tour operator UUID"
//	@Param			body	body		touroperator.Input	true	"Tour operator data"
//	@Success		200		{object}	response.CustomApiResponse{data=touroperator.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	operator, err := Update(uid, input)
	if err != nil {
		respondError(c, err, "Tour operator or geography not found", "Failed to update tour operator")
		return
	}
	response.Success(c, http.StatusOK, "Tour operator updated", operator)
}

// Delete godoc
//
//	@Summary		Delete tour operator
//	@Description	Soft-delete a tour operator by UUID
//	@Tags			Tour Operators
//	@Produce		json
//	@Param			uuid	path		string	true	"Tour operator UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	if err := Delete(uid); err != nil {
		respondError(c, err, "Tour operator not found", "Failed to delete tour operator")
		return
	}
	response.Success(c, http.StatusOK, "Tour operator deleted", nil)
}

// Users godoc
//
//	@Summary		List tour operator users
//	@Description	Staff accounts of the tour operator. Only its own staff or a platform administrator may list them.
//	@Tags			Tour Operators
//	@Produce		json
//	@Param			uuid		path		string	true	"Tour operator UUID"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name or email"
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[user.Response]}
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		403			{object}	response.CustomApiResponse
//	@Failure		404			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid}/staff [get]
func (controller Controller) Users(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	staff, err := GetStaff(uid, requesterUUID(c), pagination.FromContext(c))
	if err != nil {
		respondError(c, err, "Tour operator not found", "Failed to fetch tour operator users")
		return
	}
	response.Success(c, http.StatusOK, "Tour operator users retrieved", staff)
}

// CreateUser godoc
//
//	@Summary		Create tour operator user
//	@Description	Create a user account and link it to the tour operator. Operator staff may only grant roles they hold themselves.
//	@Tags			Tour Operators
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string		true	"Tour operator UUID"
//	@Param			body	body		user.Input	true	"User data"
//	@Success		201		{object}	response.CustomApiResponse{data=user.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid}/staff [post]
func (controller Controller) CreateUser(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	var input user.Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	created, err := CreateStaff(uid, requesterUUID(c), input)
	if err != nil {
		respondError(c, err, "Tour operator not found", "Failed to create tour operator user")
		return
	}
	response.Success(c, http.StatusCreated, "Tour operator user created", created)
}

// RemoveUser godoc
//
//	@Summary		Remove tour operator user
//	@Description	Unlink a user from the tour operator. The user account itself is kept.
//	@Tags			Tour Operators
//	@Produce		json
//	@Param			uuid		path		string	true	"Tour operator UUID"
//	@Param			user_uuid	path		string	true	"User UUID"
//	@Success		200			{object}	response.CustomApiResponse
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		403			{object}	response.CustomApiResponse
//	@Failure		404			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/tour-operators/{uuid}/staff/{user_uuid} [delete]
func (controller Controller) RemoveUser(c *gin.Context) {
	uid, ok := parseUUID(c, "uuid", "Invalid tour operator UUID")
	if !ok {
		return
	}
	userUID, ok := parseUUID(c, "user_uuid", "Invalid user UUID")
	if !ok {
		return
	}
	if err := RemoveStaff(uid, requesterUUID(c), userUID); err != nil {
		respondError(c, err, "User is not linked to this tour operator", "Failed to remove tour operator user")
		return
	}
	response.Success(c, http.StatusOK, "Tour operator user removed", nil)
}

func parseUUID(c *gin.Context, param, message string) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, message, nil)
		return uuid.Nil, false
	}
	return uid, true
}

// requesterUUID is the signed-in user, set by the Auth() middleware.
func requesterUUID(c *gin.Context) uuid.UUID {
	if v, ok := c.Get("userUUID"); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func respondError(c *gin.Context, err error, notFound, failed string) {
	switch {
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrRoleNotAllowed):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, notFound, nil)
	default:
		response.Error(c, http.StatusInternalServerError, failed)
	}
}
