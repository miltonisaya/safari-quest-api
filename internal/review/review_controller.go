package review

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

// PackageReviews godoc
//
//	@Summary		Reviews of a package
//	@Description	Visible reviews of one package with its average rating and count
//	@Tags			Reviews
//	@Produce		json
//	@Param			package_uuid	query		string	true	"Package UUID"
//	@Param			page			query		int		false	"Page number"		default(1)
//	@Param			per_page		query		int		false	"Items per page"	default(15)
//	@Param			sort_by			query		string	false	"rating, created_at or updated_at (newest first when omitted)"
//	@Param			sort_dir		query		string	false	"asc or desc"
//	@Success		200				{object}	response.CustomApiResponse{data=review.PackageList}
//	@Failure		400				{object}	response.CustomApiResponse
//	@Failure		500				{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/package-reviews [get]
func (controller Controller) PackageReviews(c *gin.Context) {
	pkg, err := uuid.Parse(c.Query("package_uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "package_uuid is required", nil)
		return
	}
	list, err := ForPackage(pkg, pagination.FromContext(c))
	if err != nil {
		fail(c, err, "Failed to fetch reviews")
		return
	}
	response.Success(c, http.StatusOK, "Reviews retrieved", list)
}

// Mine godoc
//
//	@Summary		List my reviews
//	@Tags			My Reviews
//	@Produce		json
//	@Param			page		query		int	false	"Page number"		default(1)
//	@Param			per_page	query		int	false	"Items per page"	default(15)
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[review.Response]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-reviews [get]
func (controller Controller) Mine(c *gin.Context) {
	items, err := Mine(requester(c), pagination.FromContext(c))
	if err != nil {
		fail(c, err, "Failed to fetch your reviews")
		return
	}
	response.Success(c, http.StatusOK, "Your reviews retrieved", items)
}

// Create godoc
//
//	@Summary		Review a trip
//	@Description	Review one of your own completed bookings (once per booking)
//	@Tags			My Reviews
//	@Accept			json
//	@Produce		json
//	@Param			body	body		review.Input	true	"Review"
//	@Success		201		{object}	response.CustomApiResponse{data=review.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-reviews [post]
func (controller Controller) Create(c *gin.Context) {
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Create(requester(c), input)
	if err != nil {
		fail(c, err, "Failed to save the review")
		return
	}
	response.Success(c, http.StatusCreated, "Thanks for your review", item)
}

// Update godoc
//
//	@Summary		Edit my review
//	@Tags			My Reviews
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Review UUID"
//	@Param			body	body		review.UpdateInput	true	"Review changes"
//	@Success		200		{object}	response.CustomApiResponse{data=review.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-reviews/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := UpdateMine(requester(c), uid, input)
	if err != nil {
		fail(c, err, "Failed to update the review")
		return
	}
	response.Success(c, http.StatusOK, "Review updated", item)
}

// Delete godoc
//
//	@Summary		Delete my review
//	@Tags			My Reviews
//	@Produce		json
//	@Param			uuid	path		string	true	"Review UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-reviews/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	if err := DeleteMine(requester(c), uid); err != nil {
		fail(c, err, "Failed to delete the review")
		return
	}
	response.Success(c, http.StatusOK, "Review deleted", nil)
}

// Index godoc
//
//	@Summary		Reviews of my company's packages
//	@Description	Company staff see their own packages' reviews, hidden ones included; administrators see all
//	@Tags			Reviews
//	@Produce		json
//	@Param			package_uuid	query		string	false	"Only this package"
//	@Param			page			query		int		false	"Page number"		default(1)
//	@Param			per_page		query		int		false	"Items per page"	default(15)
//	@Param			search			query		string	false	"Search title or comment"
//	@Success		200				{object}	response.CustomApiResponse{data=pagination.Result[review.Response]}
//	@Failure		400				{object}	response.CustomApiResponse
//	@Failure		403				{object}	response.CustomApiResponse
//	@Failure		500				{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/reviews [get]
func (controller Controller) Index(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to fetch reviews")
		return
	}
	var pkg *uuid.UUID
	if v := c.Query("package_uuid"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "Invalid package_uuid", nil)
			return
		}
		pkg = &id
	}
	items, err := ForCompanies(scope, pkg, pagination.FromContext(c))
	if err != nil {
		fail(c, err, "Failed to fetch reviews")
		return
	}
	response.Success(c, http.StatusOK, "Reviews retrieved", items)
}

// Reply godoc
//
//	@Summary		Reply to a review
//	@Description	Set or clear (empty reply) the company's public answer to a review of its package
//	@Tags			Reviews
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Review UUID"
//	@Param			body	body		review.ReplyInput	true	"Reply"
//	@Success		200		{object}	response.CustomApiResponse{data=review.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/review-replies/{uuid} [put]
func (controller Controller) Reply(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to save the reply")
		return
	}
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	var input ReplyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Reply(uid, input, scope)
	if err != nil {
		fail(c, err, "Failed to save the reply")
		return
	}
	response.Success(c, http.StatusOK, "Reply saved", item)
}

// Moderate godoc
//
//	@Summary		Hide or show a review
//	@Description	Administrators hide abusive reviews; hidden reviews stay stored but leave public lists and ratings
//	@Tags			Reviews
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string					true	"Review UUID"
//	@Param			body	body		review.ModerationInput	true	"Visibility"
//	@Success		200		{object}	response.CustomApiResponse{data=review.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/review-moderation/{uuid} [put]
func (controller Controller) Moderate(c *gin.Context) {
	uid, ok := pathUUID(c)
	if !ok {
		return
	}
	var input ModerationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Moderate(uid, input)
	if err != nil {
		fail(c, err, "Failed to update the review")
		return
	}
	response.Success(c, http.StatusOK, "Review visibility updated", item)
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func requester(c *gin.Context) uuid.UUID {
	if v, ok := c.Get("userUUID"); ok {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func loadScope(c *gin.Context) (companyscope.Scope, error) {
	id := requester(c)
	if id == uuid.Nil {
		return companyscope.Scope{}, errUnauthenticated
	}
	return companyscope.Load(id)
}

func pathUUID(c *gin.Context) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid review UUID", nil)
		return uuid.Nil, false
	}
	return uid, true
}

var errUnauthenticated = errors.New("Unauthorized")

func fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "Review not found", nil)
	case errors.Is(err, ErrNotYours), errors.Is(err, ErrOutsideCompany):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, companyscope.ErrNoCompany):
		response.Fail(c, http.StatusForbidden, ErrOutsideCompany.Error(), nil)
	case errors.Is(err, ErrAlreadyReviewed), errors.Is(err, ErrNotCompleted):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrInvalidRating):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, errUnauthenticated):
		response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
	default:
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
