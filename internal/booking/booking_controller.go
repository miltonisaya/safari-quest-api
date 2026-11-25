package booking

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

// ── Company staff ───────────────────────────────────────────────────────────

// Index godoc
//
//	@Summary		List bookings
//	@Description	Bookings for the caller's companies (all companies for administrators)
//	@Tags			Bookings
//	@Produce		json
//	@Param			page			query		int		false	"Page number"		default(1)
//	@Param			per_page		query		int		false	"Items per page"	default(15)
//	@Param			search			query		string	false	"Search by booking code, package or traveler"
//	@Param			sort_by			query		string	false	"Sort column: code, pax, total_price, status, created_at, updated_at"	default(created_at)
//	@Param			sort_dir		query		string	false	"Sort direction: asc or desc"	default(asc)
//	@Param			status			query		string	false	"pending, confirmed, completed or cancelled"
//	@Param			package_uuid	query		string	false	"Only bookings on this package"
//	@Success		200				{object}	response.CustomApiResponse{data=pagination.Result[booking.Response]}
//	@Failure		400				{object}	response.CustomApiResponse
//	@Failure		403				{object}	response.CustomApiResponse
//	@Failure		500				{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookings [get]
func (controller Controller) Index(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to fetch bookings")
		return
	}
	status, pkg, ok := listFilters(c)
	if !ok {
		return
	}
	items, err := GetAll(pagination.FromContext(c), status, pkg, scope)
	if err != nil {
		fail(c, err, "Failed to fetch bookings")
		return
	}
	response.Success(c, http.StatusOK, "Bookings retrieved", items)
}

// Show godoc
//
//	@Summary		Get booking
//	@Description	A single booking in one of the caller's companies
//	@Tags			Bookings
//	@Produce		json
//	@Param			uuid	path		string	true	"Booking UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=booking.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookings/{uuid} [get]
func (controller Controller) Show(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to fetch booking")
	if !ok {
		return
	}
	item, err := GetByUUID(uid, scope)
	if err != nil {
		fail(c, err, "Failed to fetch booking")
		return
	}
	response.Success(c, http.StatusOK, "Booking retrieved", item)
}

// Create godoc
//
//	@Summary		Record booking
//	@Description	Book a traveler onto one of the caller's packages. Total = price per person × pax; the package's group size cannot be exceeded.
//	@Tags			Bookings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		booking.StaffInput	true	"Booking data"
//	@Success		201		{object}	response.CustomApiResponse{data=booking.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookings [post]
func (controller Controller) Create(c *gin.Context) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, "Failed to create booking")
		return
	}
	var input StaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Create(input, scope)
	if err != nil {
		fail(c, err, "Failed to create booking")
		return
	}
	response.Success(c, http.StatusCreated, "Booking created", item)
}

// Update godoc
//
//	@Summary		Update booking
//	@Description	Change guests, notes or status. Allowed status changes: pending → confirmed or cancelled; confirmed → completed or cancelled.
//	@Tags			Bookings
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path		string				true	"Booking UUID"
//	@Param			body	body		booking.UpdateInput	true	"Booking changes"
//	@Success		200		{object}	response.CustomApiResponse{data=booking.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookings/{uuid} [put]
func (controller Controller) Update(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to update booking")
	if !ok {
		return
	}
	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Update(uid, input, scope)
	if err != nil {
		fail(c, err, "Failed to update booking")
		return
	}
	response.Success(c, http.StatusOK, "Booking updated", item)
}

// Delete godoc
//
//	@Summary		Delete booking
//	@Description	Soft-delete a booking record. To release seats normally, cancel it instead.
//	@Tags			Bookings
//	@Produce		json
//	@Param			uuid	path		string	true	"Booking UUID"
//	@Success		200		{object}	response.CustomApiResponse
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookings/{uuid} [delete]
func (controller Controller) Delete(c *gin.Context) {
	scope, uid, ok := scopeAndID(c, "Failed to delete booking")
	if !ok {
		return
	}
	if err := Delete(uid, scope); err != nil {
		fail(c, err, "Failed to delete booking")
		return
	}
	response.Success(c, http.StatusOK, "Booking deleted", nil)
}

// Travelers godoc
//
//	@Summary		Search travelers
//	@Description	Active accounts holding the Traveler role, for recording a booking on their behalf (max 20)
//	@Tags			Bookings
//	@Produce		json
//	@Param			search	query		string	false	"Name or email"
//	@Success		200		{object}	response.CustomApiResponse{data=[]booking.TravelerRef}
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/travelers [get]
func (controller Controller) Travelers(c *gin.Context) {
	items, err := Travelers(c.Query("search"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search travelers")
		return
	}
	response.Success(c, http.StatusOK, "Travelers retrieved", items)
}

// ── Travelers ───────────────────────────────────────────────────────────────

// Bookable godoc
//
//	@Summary		List bookable packages
//	@Description	Packages that have not started and still have seats, soonest first
//	@Tags			My Bookings
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			search		query		string	false	"Search by name, description or type"
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[booking.BookableResponse]}
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookable-packages [get]
func (controller Controller) Bookable(c *gin.Context) {
	items, err := Bookable(pagination.FromContext(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch packages")
		return
	}
	response.Success(c, http.StatusOK, "Bookable packages retrieved", items)
}

// BookableDetail godoc
//
//	@Summary		Package details for travelers
//	@Description	Full package information, itinerary, rating and seats left; bookable is false once it has started or is full
//	@Tags			My Bookings
//	@Produce		json
//	@Param			uuid	path		string	true	"Package UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=booking.BookableDetail}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/bookable-packages/{uuid} [get]
func (controller Controller) BookableDetail(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid package UUID", nil)
		return
	}
	item, err := BookableByUUID(uid)
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

// Mine godoc
//
//	@Summary		List my bookings
//	@Description	The signed-in traveler's own bookings
//	@Tags			My Bookings
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(15)
//	@Param			status		query		string	false	"pending, confirmed, completed or cancelled"
//	@Success		200			{object}	response.CustomApiResponse{data=pagination.Result[booking.Response]}
//	@Failure		400			{object}	response.CustomApiResponse
//	@Failure		500			{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-bookings [get]
func (controller Controller) Mine(c *gin.Context) {
	status, _, ok := listFilters(c)
	if !ok {
		return
	}
	items, err := MyBookings(pagination.FromContext(c), status, requester(c))
	if err != nil {
		fail(c, err, "Failed to fetch your bookings")
		return
	}
	response.Success(c, http.StatusOK, "Your bookings retrieved", items)
}

// Book godoc
//
//	@Summary		Book a package
//	@Description	Book seats for yourself. The booking starts as pending until the company confirms it.
//	@Tags			My Bookings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		booking.TravelerInput	true	"Package and guests"
//	@Success		201		{object}	response.CustomApiResponse{data=booking.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-bookings [post]
func (controller Controller) Book(c *gin.Context) {
	var input TravelerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, http.StatusBadRequest, "Validation failed", gin.H{"error": err.Error()})
		return
	}
	item, err := Book(requester(c), input)
	if err != nil {
		fail(c, err, "Failed to book the package")
		return
	}
	response.Success(c, http.StatusCreated, "Booking received", item)
}

// Cancel godoc
//
//	@Summary		Cancel my booking
//	@Description	Cancel one of your own pending or confirmed bookings. The record is kept with status cancelled and its seats are released.
//	@Tags			My Bookings
//	@Produce		json
//	@Param			uuid	path		string	true	"Booking UUID"
//	@Success		200		{object}	response.CustomApiResponse{data=booking.Response}
//	@Failure		400		{object}	response.CustomApiResponse
//	@Failure		403		{object}	response.CustomApiResponse
//	@Failure		404		{object}	response.CustomApiResponse
//	@Failure		409		{object}	response.CustomApiResponse
//	@Failure		500		{object}	response.CustomApiResponse
//	@Security		BearerAuth
//	@Router			/my-bookings/{uuid} [delete]
func (controller Controller) Cancel(c *gin.Context) {
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid booking UUID", nil)
		return
	}
	item, err := CancelMine(requester(c), uid)
	if err != nil {
		fail(c, err, "Failed to cancel the booking")
		return
	}
	response.Success(c, http.StatusOK, "Booking cancelled", item)
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

func scopeAndID(c *gin.Context, fallback string) (companyscope.Scope, uuid.UUID, bool) {
	scope, err := loadScope(c)
	if err != nil {
		fail(c, err, fallback)
		return companyscope.Scope{}, uuid.Nil, false
	}
	uid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid booking UUID", nil)
		return companyscope.Scope{}, uuid.Nil, false
	}
	return scope, uid, true
}

// listFilters reads ?status= and ?package_uuid=, answering 400 when invalid.
func listFilters(c *gin.Context) (Status, *uuid.UUID, bool) {
	status := Status(c.Query("status"))
	if status != "" && !ValidStatus(status) {
		response.Fail(c, http.StatusBadRequest, "Invalid status", nil)
		return "", nil, false
	}
	var pkg *uuid.UUID
	if v := c.Query("package_uuid"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "Invalid package_uuid", nil)
			return "", nil, false
		}
		pkg = &id
	}
	return status, pkg, true
}

var errUnauthenticated = errors.New("Unauthorized")

func fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Fail(c, http.StatusNotFound, "Booking not found", nil)
	case errors.Is(err, ErrOutsideCompany), errors.Is(err, companyscope.ErrNoCompany):
		response.Fail(c, http.StatusForbidden, ErrOutsideCompany.Error(), nil)
	case errors.Is(err, ErrNotYours):
		response.Fail(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, ErrNotEnoughSeats), errors.Is(err, ErrDeparted), errors.Is(err, ErrFinal), errors.Is(err, ErrUnavailable):
		response.Fail(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrInvalidReference), errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrInactiveTraveler):
		response.Fail(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, errUnauthenticated):
		response.Fail(c, http.StatusUnauthorized, "Unauthorized", nil)
	default:
		response.Error(c, http.StatusInternalServerError, fallback)
	}
}
