package booking

import (
	"errors"
	"safari-quest-api/pkg/dbtypes"
	"strings"
	"time"

	"safari-quest-api/database"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotEnoughSeats is returned when the package cannot take the requested guests.
var ErrNotEnoughSeats = errors.New("Not enough seats left on this package")

// TravelerRoleCode identifies accounts that book trips for themselves.
const TravelerRoleCode = "TRAVELER"

var sortColumns = map[string]string{
	"code":        "bookings.code",
	"pax":         "bookings.pax",
	"total_price": "bookings.total_price",
	"status":      "bookings.status",
	"created_at":  "bookings.created_at",
	"updated_at":  "bookings.updated_at",
}

// Filters narrow a booking list; empty fields are ignored.
type Filters struct {
	Status      Status
	PackageUUID *uuid.UUID
	// Scope: Admin sees everything, otherwise OperatorIDs limits the companies.
	Admin       bool
	OperatorIDs []uint
	// UserID, when set, limits the list to one traveler's bookings.
	UserID *uint
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Package").Preload("TourOperator").Preload("User")
}

func FindAll(params pagination.Params, f Filters) ([]Booking, int64, error) {
	var items []Booking
	var total int64

	db := database.GORM_DB.Model(&Booking{})
	if f.UserID != nil {
		db = db.Where("bookings.user_id = ?", *f.UserID)
	} else if !f.Admin {
		if len(f.OperatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			// By the package's current company, so bookings follow a package moved to another company.
			db = db.Where("bookings.package_id IN (SELECT id FROM packages WHERE operator_id IN ?)", f.OperatorIDs)
		}
	}
	if f.Status != "" {
		db = db.Where("bookings.status = ?", f.Status)
	}
	if f.PackageUUID != nil {
		db = db.Where("bookings.package_id = (SELECT id FROM packages WHERE uuid = ?)", *f.PackageUUID)
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Joins("JOIN users ON users.id = bookings.user_id").
			Joins("JOIN packages ON packages.id = bookings.package_id").
			Where("bookings.code ILIKE ? OR packages.name ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?",
				term, term, term, term, term)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "bookings.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (Booking, error) {
	var item Booking
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

func CodeExists(code string) (bool, error) {
	var count int64
	err := database.GORM_DB.Unscoped().Model(&Booking{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// groupRow is one group's active bookings, as read from the database.
type groupRow struct {
	PackageID  uint
	UnitNumber int
	Pax        int
	Private    bool
}

// activeGroups sums pending and confirmed bookings per package and group,
// leaving out excludeID (the booking being edited).
func activeGroups(tx *gorm.DB, packageIDs []uint, excludeID uint) ([]groupRow, error) {
	var rows []groupRow
	err := tx.Model(&Booking{}).
		Select("package_id, unit_number, COALESCE(SUM(pax), 0) AS pax, BOOL_OR(booking_type = ?) AS private", TypePrivate).
		Where("package_id IN ? AND id <> ? AND status IN ?", packageIDs, excludeID, []Status{StatusPending, StatusConfirmed}).
		Group("package_id, unit_number").Scan(&rows).Error
	return rows, err
}

func usageOf(rows []groupRow, packageID uint) usage {
	u := usage{}
	for _, r := range rows {
		if r.PackageID == packageID {
			u[r.UnitNumber] = groupUse{Held: r.Pax, Private: r.Private}
		}
	}
	return u
}

// saveChecked writes the booking inside a transaction that locks the package
// row, so concurrent bookings cannot oversell it. Seats are checked only when
// the booking holds seats after the change.
func saveChecked(b *Booking, create bool) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var pkg tourpackage.Package
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pkg, b.PackageID).Error; err != nil {
			return err
		}
		if b.Status.HoldsSeats() {
			rows, err := activeGroups(tx, []uint{pkg.ID}, b.ID)
			if err != nil {
				return err
			}
			others := usageOf(rows, pkg.ID)
			if create {
				// The group is chosen here, under the lock, so two bookings can't take the same seats.
				n, err := assignGroup(pkg, b.BookingType, b.Pax, others)
				if err != nil {
					return err
				}
				b.UnitNumber = n
			} else if !fitsGroup(pkg, b.BookingType, b.UnitNumber, b.Pax, others) {
				return ErrNotEnoughSeats
			}
		}
		omit := []string{"Package", "TourOperator", "User", "Travellers"}
		if create {
			return tx.Omit(omit...).Create(b).Error
		}
		return tx.Omit(omit...).Save(b).Error
	})
}

func Insert(b *Booking) error { return saveChecked(b, true) }

func Save(b *Booking) error { return saveChecked(b, false) }

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&Booking{}).Error
}

// BookablePackage is a package that has not started and still has room.
type BookablePackage struct {
	tourpackage.Package
	Availability Availability
}

// FindBookable lists packages starting today or later with seats left.
func FindBookable(params pagination.Params) ([]BookablePackage, int64, error) {
	// Seats used per package; a private group counts as full.
	groups := database.GORM_DB.Model(&Booking{}).
		Select("package_id, unit_number, SUM(pax) AS pax, BOOL_OR(booking_type = ?) AS private", TypePrivate).
		Where("status IN ?", []Status{StatusPending, StatusConfirmed}).
		Group("package_id, unit_number")
	used := database.GORM_DB.Table("(?) AS g", groups).
		Joins("JOIN packages p ON p.id = g.package_id").
		Select("g.package_id, SUM(CASE WHEN g.private THEN p.pax ELSE g.pax END) AS seats").
		Group("g.package_id")

	db := database.GORM_DB.Model(&tourpackage.Package{}).
		Joins("LEFT JOIN (?) AS used ON used.package_id = packages.id", used).
		Where("packages.start_date >= ?", time.Now().Format("2006-01-02")).
		Where(tourpackage.PublishedOperator).
		Where("packages.units * packages.pax - COALESCE(used.seats, 0) > 0")
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("packages.name ILIKE ? OR packages.description ILIKE ? OR packages.package_type ILIKE ?", term, term, term)
	}

	var total int64
	db.Count(&total)

	var packages []tourpackage.Package
	if err := db.Preload("Operator").
		Order("packages.start_date ASC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&packages).Error; err != nil {
		return nil, 0, err
	}

	available, err := AvailabilityFor(packages)
	if err != nil {
		return nil, 0, err
	}
	items := make([]BookablePackage, len(packages))
	for i, p := range packages {
		items[i] = BookablePackage{Package: p, Availability: available[p.ID]}
	}
	return items, total, nil
}

// FindTravelers searches active accounts holding the Traveler role.
func FindTravelers(search string, limit int) ([]user.User, error) {
	db := database.GORM_DB.Model(&user.User{}).
		Joins("JOIN user_roles ur ON ur.user_id = users.id").
		Joins("JOIN roles r ON r.id = ur.role_id AND r.code = ?", TravelerRoleCode).
		Where("users.is_active = ?", true)
	if search != "" {
		term := "%" + search + "%"
		db = db.Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", term, term, term)
	}
	var travelers []user.User
	err := db.Order("users.first_name, users.last_name").Limit(limit).Find(&travelers).Error
	return travelers, err
}

// AvailabilityFor works out what is still on sale for each package.
func AvailabilityFor(packages []tourpackage.Package) (map[uint]Availability, error) {
	out := make(map[uint]Availability, len(packages))
	if len(packages) == 0 {
		return out, nil
	}
	ids := make([]uint, len(packages))
	for i, p := range packages {
		ids[i] = p.ID
	}
	rows, err := activeGroups(database.GORM_DB, ids, 0)
	if err != nil {
		return nil, err
	}
	for _, p := range packages {
		out[p.ID] = availabilityOf(p, usageOf(rows, p.ID))
	}
	return out, nil
}

// findOverdue lists pending bookings whose confirm-by time is at or before now.
func findOverdue(now time.Time) ([]Booking, error) {
	var items []Booking
	err := withRelations(database.GORM_DB).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", StatusPending, now).
		Order("expires_at").Find(&items).Error
	return items, err
}

// staffEmails lists the active staff of a company, for booking notifications.
func staffEmails(operatorID uint) ([]string, error) {
	var emails []string
	err := database.GORM_DB.Table("users").
		Joins("JOIN tour_operator_users tou ON tou.user_id = users.id").
		Where("tou.tour_operator_id = ? AND users.is_active = ? AND users.deleted_at IS NULL", operatorID, true).
		Pluck("users.email", &emails).Error
	return emails, err
}

// ── Traveller details ──────────────────────────────────────────────────────

func travellersOf(bookingID uint) ([]Traveller, error) {
	var rows []Traveller
	err := database.GORM_DB.Where("booking_id = ?", bookingID).Order("position").Find(&rows).Error
	return rows, err
}

func leadOf(bookingID uint) (Traveller, error) {
	var t Traveller
	err := database.GORM_DB.Where("booking_id = ? AND is_lead", bookingID).Take(&t).Error
	return t, err
}

// replaceTravellers swaps the booking's travellers for rows in one
// transaction. The lead row is updated in place so an uploaded ID copy stays.
func replaceTravellers(bookingID uint, rows []Traveller, langs dbtypes.StringList, requests *string, at time.Time) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var lead Traveller
		hasLead := tx.Where("booking_id = ? AND is_lead", bookingID).Take(&lead).Error == nil
		if err := tx.Where("booking_id = ? AND NOT is_lead", bookingID).Delete(&Traveller{}).Error; err != nil {
			return err
		}
		for i := range rows {
			r := rows[i]
			r.BookingID = bookingID
			if r.IsLead && hasLead {
				r.ID, r.UUID, r.CreatedAt = lead.ID, lead.UUID, lead.CreatedAt
				r.IDDocumentPath, r.IDDocumentName, r.IDDocumentType = lead.IDDocumentPath, lead.IDDocumentName, lead.IDDocumentType
				r.IDUploadedAt, r.IDDeletedAt = lead.IDUploadedAt, lead.IDDeletedAt
				if err := tx.Save(&r).Error; err != nil {
					return err
				}
				continue
			}
			r.UUID = uuid.New()
			if err := tx.Create(&r).Error; err != nil {
				return err
			}
		}
		return tx.Model(&Booking{}).Where("id = ?", bookingID).Updates(map[string]any{
			"preferred_languages": langs, "special_requests": requests, "details_submitted_at": at,
		}).Error
	})
}

func setDocument(travellerID uint, path, name, kind *string, at time.Time) error {
	return database.GORM_DB.Model(&Traveller{}).Where("id = ?", travellerID).Updates(map[string]any{
		"id_document_path": path, "id_document_name": name, "id_document_type": kind,
		"id_uploaded_at": at, "id_deleted_at": nil,
	}).Error
}

func clearDocument(travellerID uint, at time.Time) error {
	return database.GORM_DB.Model(&Traveller{}).Where("id = ?", travellerID).Updates(map[string]any{
		"id_document_path": nil, "id_document_name": nil, "id_document_type": nil, "id_deleted_at": at,
	}).Error
}

// idsEndedBefore finds ID copies on trips that ended before cutoff, including
// soft-deleted bookings, whose files must go too.
func idsEndedBefore(cutoff time.Time) ([]Traveller, error) {
	var rows []Traveller
	err := database.GORM_DB.Raw(`
		SELECT bt.* FROM booking_travellers bt
		JOIN bookings b ON b.id = bt.booking_id
		JOIN packages p ON p.id = b.package_id
		WHERE bt.id_document_path IS NOT NULL AND p.end_date < ?`, cutoff.Format("2006-01-02")).Scan(&rows).Error
	return rows, err
}

// ── Chat ────────────────────────────────────────────────────────────────────

// participants lists who may follow a booking live: its traveler, the active
// staff of the trip's company and administrators.
func participants(bookingUUID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := database.GORM_DB.Raw(`
		SELECT u.uuid FROM bookings b JOIN users u ON u.id = b.user_id WHERE b.uuid = ?
		UNION
		SELECT u.uuid FROM bookings b
		JOIN packages p ON p.id = b.package_id
		JOIN tour_operator_users tou ON tou.tour_operator_id = p.operator_id
		JOIN users u ON u.id = tou.user_id
		WHERE b.uuid = ? AND u.is_active AND u.deleted_at IS NULL
		UNION
		SELECT u.uuid FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE r.code = 'ADMINISTRATOR' AND u.is_active AND u.deleted_at IS NULL`, bookingUUID, bookingUUID).Scan(&ids).Error
	return ids, err
}

func insertMessage(m *Message) error { return database.GORM_DB.Create(m).Error }

// markRead marks messages sent by one side as read, returning how many changed.
func markRead(bookingID uint, fromCompany bool, at time.Time) (int64, error) {
	res := database.GORM_DB.Model(&Message{}).
		Where("booking_id = ? AND from_company = ? AND read_at IS NULL", bookingID, fromCompany).
		Update("read_at", at)
	return res.RowsAffected, res.Error
}

// messagesOf returns a booking's messages, oldest first, with sender names.
func messagesOf(bookingID uint) ([]Message, map[uint]string, error) {
	var msgs []Message
	if err := database.GORM_DB.Where("booking_id = ?", bookingID).Order("created_at, id").Find(&msgs).Error; err != nil {
		return nil, nil, err
	}
	ids := []uint{}
	for _, m := range msgs {
		if m.SenderID != nil {
			ids = append(ids, *m.SenderID)
		}
	}
	names := map[uint]string{}
	if len(ids) > 0 {
		var rows []struct {
			ID        uint
			FirstName string
			LastName  string
		}
		if err := database.GORM_DB.Table("users").Select("id, first_name, last_name").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
			return nil, nil, err
		}
		for _, r := range rows {
			names[r.ID] = strings.TrimSpace(r.FirstName + " " + r.LastName)
		}
	}
	return msgs, names, nil
}

type threadRow struct {
	Booking    Booking
	Last       Message
	LastSender string
	Unread     int
	Count      int
}

// threadsFor lists bookings with messages that the viewer takes part in: as
// the traveler, or as staff of the trip's company (administrators: all).
func threadsFor(v Viewer) ([]threadRow, error) {
	var stats []struct {
		BookingID uint
		Count     int
		Unread    int
		LastID    uint
	}
	var traveler struct{ ID uint }
	if err := database.GORM_DB.Table("users").Select("id").Where("uuid = ?", v.UserUUID).Scan(&traveler).Error; err != nil {
		return nil, err
	}
	q := database.GORM_DB.Table("booking_messages m").
		Joins("JOIN bookings b ON b.id = m.booking_id AND b.deleted_at IS NULL").
		Joins("JOIN packages p ON p.id = b.package_id").
		Select(`m.booking_id,
			COUNT(*) AS count,
			COUNT(*) FILTER (WHERE m.read_at IS NULL AND m.from_company = (b.user_id = ?)) AS unread,
			MAX(m.id) AS last_id`, traveler.ID).
		Group("m.booking_id")
	switch {
	case v.Scope.Admin:
	case len(v.Scope.OperatorIDs) > 0:
		q = q.Where("b.user_id = ? OR p.operator_id IN ?", traveler.ID, v.Scope.OperatorIDs)
	default:
		q = q.Where("b.user_id = ?", traveler.ID)
	}
	if err := q.Order("last_id DESC").Limit(100).Scan(&stats).Error; err != nil {
		return nil, err
	}
	out := make([]threadRow, 0, len(stats))
	for _, st := range stats {
		var b Booking
		if err := withRelations(database.GORM_DB).First(&b, st.BookingID).Error; err != nil {
			continue
		}
		var last Message
		if err := database.GORM_DB.First(&last, st.LastID).Error; err != nil {
			continue
		}
		sender := "SafariQuest"
		if last.SenderID != nil {
			var u struct{ FirstName, LastName string }
			database.GORM_DB.Table("users").Select("first_name, last_name").Where("id = ?", *last.SenderID).Scan(&u)
			sender = strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
		out = append(out, threadRow{Booking: b, Last: last, LastSender: sender, Unread: st.Unread, Count: st.Count})
	}
	return out, nil
}
