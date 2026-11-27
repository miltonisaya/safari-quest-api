package seeders

import (
	"log"

	"safari-quest-api/database"
	"safari-quest-api/internal/authority"
	"safari-quest-api/internal/role"

	"gorm.io/gorm"
)

// travelerCodes are the only authorities a traveler needs: load their
// navigation, browse bookable packages, and manage their own bookings.
var travelerCodes = []string{
	"menu_get",
	"bookable-package_get",
	"my-booking_get", "my-booking_post", "my-booking_delete",
	"booking-traveller_get", "booking-traveller_put", "booking-traveller-id_post", "booking-traveller-id_get",
	"booking-message_get", "booking-message_post", "message-thread_get",
	"package-review_get",
	"my-review_get", "my-review_post", "my-review_put", "my-review_delete",
}

// ensureTravelerRole creates the Traveler role if it is missing and grants it
// any traveler authority it does not hold yet. Authorities an administrator
// added to the role by hand are left in place.
func ensureTravelerRole() {
	var traveler role.Role
	err := database.GORM_DB.Where("code = ?", "TRAVELER").First(&traveler).Error
	if err == gorm.ErrRecordNotFound {
		traveler = role.Role{Name: "Traveler", Code: "TRAVELER"}
		if err := database.GORM_DB.Create(&traveler).Error; err != nil {
			log.Printf("[seeder] could not create Traveler role: %v\n", err)
			return
		}
		log.Println("[seeder] Traveler role created")
	} else if err != nil {
		log.Printf("[seeder] could not look up Traveler role: %v\n", err)
		return
	}

	var wanted []authority.Authority
	if err := database.GORM_DB.Where("code IN ?", travelerCodes).Find(&wanted).Error; err != nil {
		log.Printf("[seeder] could not load traveler authorities: %v\n", err)
		return
	}
	var held []authority.Authority
	if err := database.GORM_DB.Model(&traveler).Association("Authorities").Find(&held); err != nil {
		log.Printf("[seeder] could not read traveler authorities: %v\n", err)
		return
	}
	have := make(map[string]bool, len(held))
	for _, a := range held {
		have[a.Code] = true
	}
	missing := make([]authority.Authority, 0)
	for _, a := range wanted {
		if !have[a.Code] {
			missing = append(missing, a)
		}
	}
	if len(missing) == 0 {
		return
	}
	if err := database.GORM_DB.Session(&gorm.Session{FullSaveAssociations: false}).
		Model(&traveler).Association("Authorities").Append(missing); err != nil {
		log.Printf("[seeder] could not grant traveler authorities: %v\n", err)
	}
}
