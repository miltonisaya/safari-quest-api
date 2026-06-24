package seeders

import (
	"log"

	"safari-quest-api/config"
	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/authority"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Run seeds the database in dependency order:
//  1. Authorities — derived automatically from the registered routes
//  2. Administrator role — granted every authority
//  3. Super admin user — assigned the Administrator role
func Run(router *gin.Engine) {
	log.Println("[seeder] starting...")
	authorities := seedAuthorities(router)
	role := seedAdminRole(authorities)
	seedSuperUser(role)
	log.Println("[seeder] done")
}

// seedAuthorities discovers authority codes by iterating all registered routes
// and deriving a code from each route's method and path pattern using the same
// logic the Authorize middleware uses. This guarantees the seeded codes always
// match what the middleware checks — no manual list to maintain.
func seedAuthorities(router *gin.Engine) []models.Authority {
	log.Println("[seeder] discovering authorities from registered routes...")

	// Deduplicate codes — different routes can derive the same code
	// (e.g. GET /roles and GET /roles/:uuid both touch the "role" resource).
	seen := make(map[string]bool)
	seeded := make([]models.Authority, 0)

	for _, route := range router.Routes() {
		code := authority.DeriveCode(route.Method, route.Path)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true

		var record models.Authority
		// FirstOrCreate finds the authority by code or creates it if absent.
		// Re-running the seeder will never produce duplicate authorities.
		database.GORM_DB.Where("code = ?", code).FirstOrCreate(&record, models.Authority{
			Name: code,
			Code: code,
		})
		seeded = append(seeded, record)
	}

	log.Printf("[seeder] %d authorities ready\n", len(seeded))
	return seeded
}

// seedAdminRole creates the Administrator role if it does not exist and syncs
// its authority assignments to the full list discovered from routes.
// Using Replace instead of Append means new authorities added by future routes
// are automatically granted to the Administrator role on the next seed run.
func seedAdminRole(authorities []models.Authority) models.Role {
	log.Println("[seeder] seeding Administrator role...")

	var role models.Role
	database.GORM_DB.Where("code = ?", "Administrator").FirstOrCreate(&role, models.Role{
		Name: "Administrator",
		Code: "Administrator",
	})

	if err := database.GORM_DB.Model(&role).Association("Authorities").Replace(authorities); err != nil {
		log.Printf("[seeder] warning: could not assign authorities to role: %v\n", err)
	}

	log.Println("[seeder] Administrator role ready")
	return role
}

// seedSuperUser creates the super admin user if they do not already exist and
// ensures the Administrator role is assigned to them. The email and default
// password are read from config so they never need to be hardcoded here.
func seedSuperUser(role models.Role) {
	log.Println("[seeder] seeding super admin user...")

	if config.App.AdminEmail == "" || config.App.AdminDefaultPassword == "" {
		log.Println("[seeder] skipping: ADMIN_EMAIL or ADMIN_DEFAULT_PASSWORD not set in .env")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(config.App.AdminDefaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[seeder] could not hash admin password: %v\n", err)
		return
	}

	var user models.User
	result := database.GORM_DB.Where("email = ?", config.App.AdminEmail).First(&user)

	if result.Error != nil {
		// User does not exist — create them with the Administrator role in one step.
		// GORM will insert into user_roles automatically via the Roles association.
		user = models.User{
			FirstName: "Super",
			LastName:  "Admin",
			Email:     config.App.AdminEmail,
			Password:  string(hashed),
			Sex:       "N/A",
			Mobile:    "N/A",
			Address:   "System",
			IsActive:  true,
			Roles:     []models.Role{role},
		}
		if err := database.GORM_DB.Create(&user).Error; err != nil {
			log.Printf("[seeder] could not create super user: %v\n", err)
			return
		}
		log.Printf("[seeder] super user created: %s\n", config.App.AdminEmail)
	} else {
		// User exists — refresh the role assignment in case it was removed.
		if err := database.GORM_DB.Model(&user).Association("Roles").Replace([]models.Role{role}); err != nil {
			log.Printf("[seeder] could not assign role to super user: %v\n", err)
		}
		log.Println("[seeder] super user already exists, role assignment refreshed")
	}
}
