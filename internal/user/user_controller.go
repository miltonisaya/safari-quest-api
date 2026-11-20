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
//	@Param			role_uuid	query		string	false	"Filter by role UUID"
//	@Param			active		query		string	false	"Filter by status: true or false"
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[Response]}
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users [get]
func (controller Controller) Index(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		failUser(c, err, "Failed to fetch users")
		return
	}
	params := pagination.FromContext(c)
	filter, err := parseListFilter(c.Query("role_uuid"), c.Query("active"))
	if err != nil {
		failUser(c, err, "Failed to fetch users")
		return
	}
	users, err := GetAll(params, scope.apply(filter))
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
	scope, err := loadScope(c)
	if err != nil {
		failUser(c, err, "Failed to fetch user")
		return
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	target, err := FindByUUID(uid)
	if err != nil {
		failUser(c, err, "Failed to fetch user")
		return
	}
	if err := scope.allow(target); err != nil {
		failUser(c, err, "Failed to fetch user")
		return
	}
	response.Success(c, http.StatusOK, "User retrieved", ToResponse(target))
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
	scope, err := loadScope(c)
	if err != nil {
		failUser(c, err, "Failed to create user")
		return
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	if err := mayGrant(scope.admin, roleUUIDs(scope.account.Roles), input.RoleIDs); err != nil {
		failUser(c, err, "Failed to create user")
		return
	}
	created, err := Create(input)
	if err != nil {
		failUser(c, err, "Failed to create user")
		return
	}
	if !scope.admin {
		account, findErr := FindByUUID(created.UUID)
		if findErr == nil {
			findErr = AttachToCompanies(account.ID, scope.operatorIDs)
		}
		if findErr != nil {
			_ = Remove(created.UUID)
			failUser(c, findErr, "Failed to create user")
			return
		}
	}
	response.Success(c, http.StatusCreated, "User created", created)
}

// Update godoc
//
//	@Summary		Update user
//	@Description	Update an existing user by UUID. is_active changes status when present. reset_password sets the password back to the system default.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string		true	"User UUID"
//	@Param			body	body		UpdateInput	true	"User data"
//	@Success		200		{object}	response.CustomApiResponse{data=Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		failUser(c, err, "Failed to update user")
		return
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	target, err := FindByUUID(uid)
	if err != nil {
		failUser(c, err, "Failed to update user")
		return
	}
	if err := scope.allow(target); err != nil {
		failUser(c, err, "Failed to update user")
		return
	}
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	if err := mayGrant(scope.admin, roleUUIDs(scope.account.Roles), input.RoleIDs); err != nil {
		failUser(c, err, "Failed to update user")
		return
	}
	user, err := Update(scope.account.UUID, uid, input)
	if err != nil {
		failUser(c, err, "Failed to update user")
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
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/users/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		failUser(c, err, "Failed to delete user")
		return
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user UUID", nil)
		return
	}
	target, err := FindByUUID(uid)
	if err != nil {
		failUser(c, err, "Failed to delete user")
		return
	}
	if err := scope.allow(target); err != nil {
		failUser(c, err, "Failed to delete user")
		return
	}
	if err := Delete(scope.account.UUID, uid); err != nil {
		failUser(c, err, "Failed to delete user")
		return
	}
	response.Success(c, http.StatusOK, "User deleted", nil)
}

type actorScope struct {
	account     User
	admin       bool
	operatorIDs []uint
}

func loadScope(c *gin.Context) (actorScope, error) {
	id, ok := actorUUID(c)
	if !ok {
		return actorScope{}, errUnauthenticated
	}
	account, err := FindByUUID(id)
	if err != nil {
		return actorScope{}, err
	}
	if isAdministrator(account.Roles) {
		return actorScope{account: account, admin: true}, nil
	}
	ids, err := OperatorIDsForUser(account.ID)
	if err != nil {
		return actorScope{}, err
	}
	if len(ids) == 0 {
		return actorScope{}, ErrOutsideCompany
	}
	return actorScope{account: account, operatorIDs: ids}, nil
}

func (scope actorScope) apply(filter listFilter) listFilter {
	if scope.admin {
		return filter
	}
	filter.CompanyOnly = true
	filter.OperatorIDs = scope.operatorIDs
	filter.HeldRoleUUIDs = roleUUIDs(scope.account.Roles)
	return filter
}

func (scope actorScope) allow(target User) error {
	if scope.admin {
		return nil
	}
	same, err := SharesCompany(scope.operatorIDs, target.ID)
	if err != nil {
		return err
	}
	return mayManage(false, same, roleUUIDs(scope.account.Roles), roleUUIDs(target.Roles))
}

func actorUUID(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get("userUUID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}

func failUser(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "User not found", nil)
	case errors.Is(err, ErrInvalidRoleFilter), errors.Is(err, ErrInvalidStatusFilter):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, ErrCannotDeactivateSelf), errors.Is(err, ErrCannotDeleteSelf):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrOutsideCompany), errors.Is(err, ErrRoleNotAllowed):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, errUnauthenticated):
		response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
	case errors.Is(err, ErrDefaultPasswordUnset):
		response.Error(c, http.StatusInternalServerError, "Default password is not configured")
	default:
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
