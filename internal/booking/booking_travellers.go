package booking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/internal/agegroup"
	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/setting"
	"safari-quest-api/pkg/dbtypes"
	"safari-quest-api/pkg/geo"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrDetailsClosed   = errors.New("Traveller details can only be changed before the trip starts, while the booking is pending or confirmed")
	ErrInvalidDetails  = errors.New("Check the traveller details")
	ErrNoLeadYet       = errors.New("Save the traveller details before uploading the ID")
	ErrInvalidDocument = errors.New("Upload a photo (JPG, PNG or WebP) or a PDF of the ID, up to 10 MB")
	ErrNoDocument      = errors.New("There is no ID copy for this booking")
)

const maxIDBytes = 10 << 20

// ── Inputs ──────────────────────────────────────────────────────────────────

type LeadInput struct {
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Nationality  string    `json:"nationality"`
	AgeGroupUUID uuid.UUID `json:"age_group_uuid"`
	IDType       string    `json:"id_type"`
	IDNumber     string    `json:"id_number"`
	IDExpiry     string    `json:"id_expiry"` // YYYY-MM-DD
}

type MemberInput struct {
	FullName     string    `json:"full_name"`
	Nationality  string    `json:"nationality"`
	AgeGroupUUID uuid.UUID `json:"age_group_uuid"`
}

// DetailsInput is the whole traveller details form; it replaces what was saved.
type DetailsInput struct {
	PreferredLanguages []string      `json:"preferred_languages"`
	SpecialRequests    *string       `json:"special_requests"`
	Lead               LeadInput     `json:"lead"`
	Members            []MemberInput `json:"members"`
}

// ── Responses ───────────────────────────────────────────────────────────────

type DocumentInfo struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	UploadedAt *time.Time `json:"uploaded_at"`
}

type TravellerResponse struct {
	UUID        uuid.UUID     `json:"uuid"`
	IsLead      bool          `json:"is_lead"`
	FullName    string        `json:"full_name"`
	Email       *string       `json:"email"`
	Phone       *string       `json:"phone"`
	Nationality string        `json:"nationality"`
	AgeGroup    string        `json:"age_group"`
	AgeGroupID  *uuid.UUID    `json:"age_group_uuid"`
	IDType      *string       `json:"id_type"`
	IDNumber    *string       `json:"id_number"`
	IDExpiry    *string       `json:"id_expiry"`
	IDDocument  *DocumentInfo `json:"id_document"` // nil when none (or deleted)
	IDDeletedAt *time.Time    `json:"id_deleted_at"`
}

type DetailsResponse struct {
	BookingUUID        uuid.UUID           `json:"booking_uuid"`
	Code               string              `json:"code"`
	Pax                int                 `json:"pax"`
	Status             Status              `json:"status"`
	Editable           bool                `json:"editable"`
	PreferredLanguages []string            `json:"preferred_languages"`
	SpecialRequests    *string             `json:"special_requests"`
	SubmittedAt        *time.Time          `json:"submitted_at"`
	Travellers         []TravellerResponse `json:"travellers"`
}

// ── Access ──────────────────────────────────────────────────────────────────

// Viewer is who is asking: the signed-in user and, for staff, their companies.
type Viewer struct {
	UserUUID uuid.UUID
	Scope    companyscope.Scope
}

// canSee: the booking's traveler, or staff of the trip's company (or an administrator).
func canSee(b Booking, v Viewer) bool {
	if b.User.UUID == v.UserUUID {
		return true
	}
	return v.Scope.Allows(b.Package.OperatorID)
}

func editable(b Booking) bool {
	return (b.Status == StatusPending || b.Status == StatusConfirmed) && !started(b.Package)
}

func ownBooking(travelerUUID, bookingUUID uuid.UUID) (Booking, error) {
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return Booking{}, err
	}
	if b.User.UUID != travelerUUID {
		return Booking{}, ErrNotYours
	}
	return b, nil
}

// ── Reading ─────────────────────────────────────────────────────────────────

func travellerResponses(rows []Traveller, groupUUIDs map[uint]uuid.UUID) []TravellerResponse {
	out := make([]TravellerResponse, len(rows))
	for i, t := range rows {
		r := TravellerResponse{
			UUID: t.UUID, IsLead: t.IsLead, FullName: t.FullName, Email: t.Email, Phone: t.Phone,
			Nationality: t.Nationality, AgeGroup: t.AgeGroupName, IDType: t.IDType, IDNumber: t.IDNumber, IDDeletedAt: t.IDDeletedAt,
		}
		if t.AgeGroupID != nil {
			if id, ok := groupUUIDs[*t.AgeGroupID]; ok {
				r.AgeGroupID = &id
			}
		}
		if t.IDExpiry != nil {
			s := t.IDExpiry.Format("2006-01-02")
			r.IDExpiry = &s
		}
		if t.IDDocumentPath != nil {
			r.IDDocument = &DocumentInfo{Name: deref(t.IDDocumentName), Type: deref(t.IDDocumentType), UploadedAt: t.IDUploadedAt}
		}
		out[i] = r
	}
	return out
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func detailsOf(b Booking) (DetailsResponse, error) {
	rows, err := travellersOf(b.ID)
	if err != nil {
		return DetailsResponse{}, err
	}
	groups, err := agegroup.FindAll()
	if err != nil {
		return DetailsResponse{}, err
	}
	ids := map[uint]uuid.UUID{}
	for _, g := range groups {
		ids[g.ID] = g.UUID
	}
	langs := []string(b.PreferredLanguages)
	if langs == nil {
		langs = []string{}
	}
	return DetailsResponse{
		BookingUUID: b.UUID, Code: b.Code, Pax: b.Pax, Status: b.Status, Editable: editable(b),
		PreferredLanguages: langs, SpecialRequests: b.SpecialRequests, SubmittedAt: b.DetailsSubmittedAt,
		Travellers: travellerResponses(rows, ids),
	}, nil
}

// Details returns a booking's traveller details to its traveler or the company.
func Details(bookingUUID uuid.UUID, v Viewer) (DetailsResponse, error) {
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return DetailsResponse{}, err
	}
	if !canSee(b, v) {
		return DetailsResponse{}, ErrNotYours
	}
	return detailsOf(b)
}

// ── Saving ──────────────────────────────────────────────────────────────────

var (
	phonePattern = regexp.MustCompile(`^\+?[\d\s()-]{7,30}$`)
	idPattern    = regexp.MustCompile(`^[A-Za-z0-9-]{3,50}$`)
)

func ageGroupSnapshot(id uuid.UUID, groups map[uuid.UUID]agegroup.AgeGroup) (agegroup.AgeGroup, bool) {
	g, ok := groups[id]
	return g, ok
}

func bad(field string) error { return fmt.Errorf("%w: %s", ErrInvalidDetails, field) }

// buildTravellers validates the form and returns the rows to store.
func buildTravellers(b Booking, in DetailsInput, groups map[uuid.UUID]agegroup.AgeGroup) ([]Traveller, []string, error) {
	langs := make([]string, 0, len(in.PreferredLanguages))
	seen := map[string]bool{}
	for _, l := range in.PreferredLanguages {
		l = strings.ToLower(strings.TrimSpace(l))
		if !geo.IsLanguage(l) {
			return nil, nil, bad("preferred languages")
		}
		if !seen[l] {
			seen[l] = true
			langs = append(langs, l)
		}
	}
	if len(langs) == 0 || len(langs) > 5 {
		return nil, nil, bad("choose 1 to 5 preferred languages")
	}
	if in.SpecialRequests != nil && len(*in.SpecialRequests) > 1000 {
		return nil, nil, bad("special requests (1000 characters at most)")
	}
	if len(in.Members) != b.Pax-1 {
		return nil, nil, bad(fmt.Sprintf("this booking is for %d travellers, so add %d other group members", b.Pax, b.Pax-1))
	}

	l := in.Lead
	name := strings.TrimSpace(l.FullName)
	email := strings.TrimSpace(l.Email)
	phone := strings.TrimSpace(l.Phone)
	nat := strings.ToUpper(strings.TrimSpace(l.Nationality))
	if len(name) < 2 || len(name) > 200 {
		return nil, nil, bad("lead traveller's full name")
	}
	if _, err := mail.ParseAddress(email); err != nil || len(email) > 255 {
		return nil, nil, bad("lead traveller's email")
	}
	if !phonePattern.MatchString(phone) {
		return nil, nil, bad("lead traveller's phone")
	}
	if !geo.IsCountry(nat) {
		return nil, nil, bad("lead traveller's nationality")
	}
	g, ok := ageGroupSnapshot(l.AgeGroupUUID, groups)
	if !ok {
		return nil, nil, bad("lead traveller's age group")
	}
	if l.IDType != IDPassport && l.IDType != IDNationalID {
		return nil, nil, bad("ID type (passport or national ID)")
	}
	idNumber := strings.ToUpper(strings.TrimSpace(l.IDNumber))
	if !idPattern.MatchString(idNumber) {
		return nil, nil, bad("ID number (letters, digits and dashes)")
	}
	expiry, err := time.Parse("2006-01-02", l.IDExpiry)
	if err != nil {
		return nil, nil, bad("ID expiry date")
	}
	if expiry.Before(b.Package.EndDate) {
		return nil, nil, bad("the ID expires before the trip ends")
	}
	idType := l.IDType
	rows := []Traveller{{
		IsLead: true, Position: 1, FullName: name, Email: &email, Phone: &phone, Nationality: nat,
		AgeGroupID: &g.ID, AgeGroupName: g.Name, AgeMin: g.MinAge, AgeMax: g.MaxAge,
		IDType: &idType, IDNumber: &idNumber, IDExpiry: &expiry,
	}}
	for i, m := range in.Members {
		mName := strings.TrimSpace(m.FullName)
		mNat := strings.ToUpper(strings.TrimSpace(m.Nationality))
		label := fmt.Sprintf("group member %d", i+1)
		if len(mName) < 2 || len(mName) > 200 {
			return nil, nil, bad(label + "'s full name")
		}
		if !geo.IsCountry(mNat) {
			return nil, nil, bad(label + "'s nationality")
		}
		mg, ok := ageGroupSnapshot(m.AgeGroupUUID, groups)
		if !ok {
			return nil, nil, bad(label + "'s age group")
		}
		rows = append(rows, Traveller{
			Position: i + 2, FullName: mName, Nationality: mNat,
			AgeGroupID: &mg.ID, AgeGroupName: mg.Name, AgeMin: mg.MinAge, AgeMax: mg.MaxAge,
		})
	}
	return rows, langs, nil
}

// SaveDetails stores the traveller details of one of the traveler's bookings.
func SaveDetails(travelerUUID, bookingUUID uuid.UUID, in DetailsInput) (DetailsResponse, error) {
	b, err := ownBooking(travelerUUID, bookingUUID)
	if err != nil {
		return DetailsResponse{}, err
	}
	if !editable(b) {
		return DetailsResponse{}, ErrDetailsClosed
	}
	all, err := agegroup.FindAll()
	if err != nil {
		return DetailsResponse{}, err
	}
	groups := map[uuid.UUID]agegroup.AgeGroup{}
	for _, g := range all {
		groups[g.UUID] = g
	}
	rows, langs, err := buildTravellers(b, in, groups)
	if err != nil {
		return DetailsResponse{}, err
	}
	var requests *string
	if in.SpecialRequests != nil {
		if r := strings.TrimSpace(*in.SpecialRequests); r != "" {
			requests = &r
		}
	}
	now := time.Now()
	if err := replaceTravellers(b.ID, rows, dbtypes.StringList(langs), requests, now); err != nil {
		return DetailsResponse{}, err
	}
	saved, err := FindByUUID(b.UUID)
	if err != nil {
		return DetailsResponse{}, err
	}
	publish(ws.ActionUpdated, ToResponse(saved))
	return detailsOf(saved)
}

// ── ID copy ─────────────────────────────────────────────────────────────────

var documentTypes = map[string]string{
	"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "application/pdf": ".pdf",
}

func privatePath(rel string) string {
	return filepath.Join(config.App.PrivateUploadDir, rel)
}

// UploadID stores a copy of the lead traveller's ID, replacing any earlier one.
func UploadID(travelerUUID, bookingUUID uuid.UUID, file io.Reader, name string) (DetailsResponse, error) {
	b, err := ownBooking(travelerUUID, bookingUUID)
	if err != nil {
		return DetailsResponse{}, err
	}
	if !editable(b) {
		return DetailsResponse{}, ErrDetailsClosed
	}
	lead, err := leadOf(b.ID)
	if err != nil {
		return DetailsResponse{}, ErrNoLeadYet
	}
	data, err := io.ReadAll(io.LimitReader(file, maxIDBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxIDBytes {
		return DetailsResponse{}, ErrInvalidDocument
	}
	// Trust the bytes, not the file name.
	kind := http.DetectContentType(data)
	if kind == "application/octet-stream" && strings.HasSuffix(strings.ToLower(name), ".webp") && len(data) > 12 && string(data[8:12]) == "WEBP" {
		kind = "image/webp"
	}
	ext, ok := documentTypes[kind]
	if !ok {
		return DetailsResponse{}, ErrInvalidDocument
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return DetailsResponse{}, err
	}
	rel := filepath.Join("ids", hex.EncodeToString(random)+ext)
	if err := os.MkdirAll(filepath.Dir(privatePath(rel)), 0o700); err != nil {
		return DetailsResponse{}, err
	}
	if err := os.WriteFile(privatePath(rel), data, 0o600); err != nil {
		return DetailsResponse{}, err
	}
	old := lead.IDDocumentPath
	clean := filepath.Base(name)
	if len(clean) > 200 {
		clean = clean[:200]
	}
	if err := setDocument(lead.ID, &rel, &clean, &kind, time.Now()); err != nil {
		_ = os.Remove(privatePath(rel))
		return DetailsResponse{}, err
	}
	if old != nil {
		_ = os.Remove(privatePath(*old))
	}
	return detailsOf(b)
}

// IDDocument returns the lead's ID copy to the traveler or the trip's company.
func IDDocument(bookingUUID uuid.UUID, v Viewer) ([]byte, string, string, error) {
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return nil, "", "", err
	}
	if !canSee(b, v) {
		return nil, "", "", ErrNotYours
	}
	lead, err := leadOf(b.ID)
	if err != nil || lead.IDDocumentPath == nil {
		return nil, "", "", ErrNoDocument
	}
	data, err := os.ReadFile(privatePath(*lead.IDDocumentPath))
	if err != nil {
		return nil, "", "", ErrNoDocument
	}
	name := fmt.Sprintf("%s-ID%s", b.Code, filepath.Ext(*lead.IDDocumentPath))
	return data, deref(lead.IDDocumentType), name, nil
}

// ── Retention ───────────────────────────────────────────────────────────────

// IDCleanupEvery is how often expired ID copies are looked for.
const IDCleanupEvery = 6 * time.Hour

// DeleteExpiredIDs removes ID copies of trips that ended more than the
// configured number of days before now. The booking and traveller rows stay.
func DeleteExpiredIDs(now time.Time) (int, error) {
	days := setting.Int(setting.IDDocumentRetentionDays)
	cutoff := now.AddDate(0, 0, -days)
	rows, err := idsEndedBefore(cutoff)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, t := range rows {
		if t.IDDocumentPath != nil {
			if err := os.Remove(privatePath(*t.IDDocumentPath)); err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Printf("[booking] could not delete ID copy %d: %v\n", t.ID, err)
				continue
			}
		}
		if err := clearDocument(t.ID, now); err != nil {
			log.Printf("[booking] could not record ID deletion %d: %v\n", t.ID, err)
			continue
		}
		n++
	}
	return n, nil
}

// RunIDCleanup deletes expired ID copies now and every IDCleanupEvery.
func RunIDCleanup(ctx context.Context) {
	run := func() {
		if n, err := DeleteExpiredIDs(time.Now()); err != nil {
			log.Printf("[booking] ID clean-up failed: %v\n", err)
		} else if n > 0 {
			log.Printf("[booking] deleted %d ID copy(ies) past the retention period\n", n)
		}
	}
	run()
	ticker := time.NewTicker(IDCleanupEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

// viewerFor builds the Viewer for a signed-in user.
func viewerFor(userUUID uuid.UUID) (Viewer, error) {
	scope, err := companyscope.Load(userUUID)
	if err != nil && !errors.Is(err, companyscope.ErrNoCompany) {
		return Viewer{}, err
	}
	return Viewer{UserUUID: userUUID, Scope: scope}, nil
}
