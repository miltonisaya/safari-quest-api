// Package catalog is the public, read-only view of sale packages for the
// traveler-facing website. It needs no sign-in and never exposes company
// staff or traveler details beyond a review's first name and initial.
package catalog

import (
	"path/filepath"
	"strings"
	"time"

	"safari-quest-api/internal/booking"
	"safari-quest-api/internal/review"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MaxPerPage caps a catalog page; the website loads upcoming packages in one go.
const MaxPerPage = 100

var imageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".avif": true}

// Package is a package as the public website shows it.
type Package struct {
	tourpackage.Response
	// ImagePath is the first image attachment, relative to the API root
	// (e.g. "uploads/abc.png", served at /api/v1/uploads/abc.png); empty when
	// the package has no image.
	ImagePath            string `json:"image_path"`
	DurationDays         int    `json:"duration_days"`
	booking.Availability        // seats_left, max_shared_travellers, private_groups_left
	Bookable             bool   `json:"bookable"` // not started and something left to book
}

type List struct {
	Items []Package       `json:"items"`
	Meta  pagination.Meta `json:"meta"`
}

func imagePath(p tourpackage.Package) string {
	for _, f := range p.FileResources {
		if imageExt[strings.ToLower(filepath.Ext(f.Path))] {
			return f.Path
		}
	}
	return ""
}

func toPublic(p tourpackage.Package, rating tourpackage.Rating, a booking.Availability) Package {
	resp := tourpackage.ToResponseRated(p, rating)
	// The public site does not need attachment metadata or staff-facing paths.
	resp.FileResources = []tourpackage.FileRef{}
	today := time.Now().Format("2006-01-02")
	return Package{
		Response:     resp,
		ImagePath:    imagePath(p),
		DurationDays: int(p.EndDate.Sub(p.StartDate).Hours()/24) + 1,
		Availability: a,
		Bookable:     a.Bookable() && p.StartDate.Format("2006-01-02") >= today && tourpackage.Published(p),
	}
}

// Packages lists upcoming packages with ratings and seats left.
func Packages(f tourpackage.CatalogFilters, page, perPage int) (List, error) {
	if perPage < 1 || perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	if page < 1 {
		page = 1
	}
	rows, total, err := tourpackage.FindCatalog(f, perPage, (page-1)*perPage)
	if err != nil {
		return List{}, err
	}
	ids := make([]uint, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	ratings, err := tourpackage.RatingsFor(ids)
	if err != nil {
		return List{}, err
	}
	available, err := booking.AvailabilityFor(rows)
	if err != nil {
		return List{}, err
	}
	items := make([]Package, len(rows))
	for i, r := range rows {
		items[i] = toPublic(r, ratings[r.ID], available[r.ID])
	}
	meta := pagination.NewMeta(total, pagination.Params{Page: page, PerPage: perPage})
	return List{Items: items, Meta: meta}, nil
}

// PackageByUUID returns one package, including ones that already started
// (Bookable is then false).
func PackageByUUID(id uuid.UUID) (Package, error) {
	p, err := tourpackage.FindByUUID(id)
	if err != nil {
		return Package{}, err
	}
	if !tourpackage.Published(p) {
		// Unverified companies' packages are not public.
		return Package{}, gorm.ErrRecordNotFound
	}
	ratings, err := tourpackage.RatingsFor([]uint{p.ID})
	if err != nil {
		return Package{}, err
	}
	available, err := booking.AvailabilityFor([]tourpackage.Package{p})
	if err != nil {
		return Package{}, err
	}
	return toPublic(p, ratings[p.ID], available[p.ID]), nil
}

// Reviews lists a package's visible reviews with its rating summary.
func Reviews(id uuid.UUID, params pagination.Params) (review.PackageList, error) {
	return review.ForPackage(id, params)
}
