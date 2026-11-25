package seeders

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"safari-quest-api/database"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/pkg/dbtypes"

	"gorm.io/gorm"
)

//go:embed demo_packages.json
var demoPackagesJSON []byte

type demoDay struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Accommodation *string  `json:"accommodation"`
	Meals         []string `json:"meals"`
}

type demoPackage struct {
	Name         string    `json:"name"`
	Operator     string    `json:"operator"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	Days         int       `json:"days"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	EcoCertified bool      `json:"eco_certified"`
	PackageType  string    `json:"package_type"`
	Pax          int       `json:"pax"`
	Difficulty   *string   `json:"difficulty"`
	Highlights   []string  `json:"highlights"`
	Inclusions   []string  `json:"inclusions"`
	Exclusions   []string  `json:"exclusions"`
	Itinerary    []demoDay `json:"itinerary"`
	StartsInDays int       `json:"starts_in_days"`
}

var nonAlnum = regexp.MustCompile(`[^A-Z0-9]+`)

// demoCode turns "Baobab Safari Co." into "DEMO_BAOBAB_SAFARI_CO".
func demoCode(name string) string {
	return "DEMO_" + strings.Trim(nonAlnum.ReplaceAllString(strings.ToUpper(name), "_"), "_")
}

// SeedDemoPackages loads the website's sample tour companies and their
// packages (from demo_packages.json) as real records, starting 30+ days from
// today. Companies are matched by code and packages by name within their
// company, so running it again only adds what is missing.
func SeedDemoPackages() {
	var items []demoPackage
	if err := json.Unmarshal(demoPackagesJSON, &items); err != nil {
		log.Printf("[seeder] demo packages: %v\n", err)
		return
	}
	created := 0
	for _, item := range items {
		operator, err := ensureDemoOperator(item.Operator)
		if err != nil {
			log.Printf("[seeder] demo company %q: %v\n", item.Operator, err)
			continue
		}
		var count int64
		database.GORM_DB.Model(&tourpackage.Package{}).
			Where("operator_id = ? AND name = ?", operator.ID, item.Name).Count(&count)
		if count > 0 {
			continue
		}
		if err := insertDemoPackage(item, operator); err != nil {
			log.Printf("[seeder] demo package %q: %v\n", item.Name, err)
			continue
		}
		created++
	}
	log.Printf("[seeder] demo packages ready (%d created)\n", created)
}

func ensureDemoOperator(name string) (touroperator.TourOperator, error) {
	code := demoCode(name)
	var op touroperator.TourOperator
	err := database.GORM_DB.Where("code = ?", code).First(&op).Error
	if err == nil {
		return op, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return op, err
	}
	slug := strings.ToLower(strings.TrimPrefix(code, "DEMO_"))
	description := "Sample company loaded with SEED_DEMO=true."
	op = touroperator.TourOperator{
		Name:        name,
		Code:        code,
		Email:       fmt.Sprintf("hello@%s.example", strings.ReplaceAll(slug, "_", "-")),
		Phone:       "+255 700 000 000",
		Description: &description,
		IsVerified:  true,
		IsActive:    true,
	}
	err = touroperator.Insert(&op)
	return op, err
}

func insertDemoPackage(item demoPackage, operator touroperator.TourOperator) error {
	days := item.Days
	if days < 1 {
		days = 1
	}
	start := time.Now().AddDate(0, 0, item.StartsInDays).Truncate(24 * time.Hour)
	location := item.Location
	pkg := tourpackage.Package{
		OperatorID:   operator.ID,
		Name:         item.Name,
		Description:  item.Description,
		StartDate:    start,
		EndDate:      start.AddDate(0, 0, days-1),
		Pax:          item.Pax,
		Price:        item.Price,
		PackageType:  item.PackageType,
		Highlights:   dbtypes.StringList(item.Highlights),
		Inclusions:   dbtypes.StringList(item.Inclusions),
		Exclusions:   dbtypes.StringList(item.Exclusions),
		Difficulty:   item.Difficulty,
		Location:     &location,
		Currency:     item.Currency,
		EcoCertified: item.EcoCertified,
	}
	itinerary := make([]tourpackage.ItineraryDay, len(item.Itinerary))
	for i, d := range item.Itinerary {
		itinerary[i] = tourpackage.ItineraryDay{
			DayNumber:     i + 1,
			Title:         d.Title,
			Description:   d.Description,
			Accommodation: d.Accommodation,
			Meals:         dbtypes.StringList(d.Meals),
		}
	}
	return tourpackage.Insert(&pkg, tourpackage.Links{Itinerary: itinerary})
}
