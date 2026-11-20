package seeders

import (
	"errors"
	"log"
	"net/http"

	"safari-quest-api/database"
	"safari-quest-api/internal/authority"
	"safari-quest-api/internal/menu"
	authoritycode "safari-quest-api/pkg/authority"

	"gorm.io/gorm"
)

// menuSeed is one row in the default navigation. Groups nest their links in
// Children. Authority is an existing authority code (resource_get).
type menuSeed struct {
	Name      string
	Type      menu.Type
	Route     string
	Icon      string
	Authority string
	SortOrder int
	Children  []menuSeed
}

// appMenuTree is the signed-in navigation for an empty database. The left menu
// shows two links, then Settings and Security & Administration. Location and
// attraction pages open under Settings.
// Routes match the Angular dashboard pages. Authority codes come from the same
// derivation the Authorize middleware uses, so a menu item stays attached to
// the permission of the list endpoint behind its page.
func appMenuTree() []menuSeed {
	// shortcut: dashboard and messages have no API resource, keep them on menu_get until a traveler role exists.
	shell := authoritycode.DeriveCode(http.MethodGet, "/api/v1/menus")
	return []menuSeed{
		{Name: "Dashboard", Type: menu.TypeLink, Route: "/dashboard", Icon: "dashboard", Authority: shell, SortOrder: 10},
		{Name: "Messages", Type: menu.TypeLink, Route: "/messages", Icon: "forum", Authority: shell, SortOrder: 20},
		group("Settings", "settings", "/geographies", 40,
			group("Location Management", "location_on", "/geographies", 10,
				link("Geographies", "/geographies", "map", "/geographies", 10),
				link("Geography Types", "/geography-types", "place", "/geography-types", 20),
			),
			group("Attractions Management", "travel_explore", "/attractions", 30,
				link("Attractions", "/attractions", "park", "/attractions", 10),
				link("Attraction Types", "/attraction-types", "tag", "/attraction-types", 20),
				link("Attraction Categories", "/attraction-categories", "label", "/attraction-categories", 30),
				link("Tour Operators", "/tour-operators", "store", "/tour-operators", 40),
				link("Packages", "/packages", "inventory", "/packages", 50),
				link("File Resources", "/file-resources", "attach_file", "/file-resources", 60),
			),
		),
		group("Security & Administration", "admin_panel_settings", "/users", 50,
			link("Users", "/users", "group", "/users", 10),
			link("Roles", "/roles", "badge", "/roles", 20),
			link("Authorities", "/authorities", "verified_user", "/authorities", 30),
			link("Menus", "/menus", "menu_open", "/menus", 40),
			link("File Resource Types", "/file-resource-types", "folder_open", "/file-resource-types", 50),
		),
	}
}

func link(name, route, icon, apiPath string, sort int) menuSeed {
	return menuSeed{
		Name:      name,
		Type:      menu.TypeLink,
		Route:     route,
		Icon:      icon,
		Authority: authoritycode.DeriveCode(http.MethodGet, "/api/v1"+apiPath),
		SortOrder: sort,
	}
}

func group(name, icon, apiPath string, sort int, children ...menuSeed) menuSeed {
	return menuSeed{
		Name:      name,
		Type:      menu.TypeGroup,
		Icon:      icon,
		Authority: authoritycode.DeriveCode(http.MethodGet, "/api/v1"+apiPath),
		SortOrder: sort,
		Children:  children,
	}
}

// SeedMenus inserts the default navigation on an empty menu table. Matching is
// by name, type, and parent. Once any top-level menu exists, later boots leave
// the table alone so edits made in the Menus screen are kept. Authorities must
// already be seeded.
func SeedMenus() {
	var roots int64
	if err := database.GORM_DB.Model(&menu.Menu{}).Where("parent_id IS NULL").Count(&roots).Error; err != nil {
		log.Printf("[seeder] menu count: %v\n", err)
		return
	}
	if roots > 0 {
		log.Println("[seeder] menus already present, leaving them")
		return
	}

	log.Println("[seeder] seeding menus...")
	created := 0
	for _, item := range appMenuTree() {
		created += ensureMenu(item, nil)
	}
	log.Printf("[seeder] menus ready (%d created)\n", created)
}

// ensureFeatureMenus adds package and file links on a database that already
// has a menu. Existing rows are left as they are.
// shortcut: a deleted link with the same route is inserted again on the next boot, stop if menu deletes must stick.
func ensureFeatureMenus() {
	packagesParent := menuIDByName("Attractions Management")
	if packagesParent == nil {
		packagesParent = menuIDByName("Settings")
	}
	ensureLink(packagesParent, "Packages", "/packages", "inventory", "/packages", 50)
	ensureLink(packagesParent, "File Resources", "/file-resources", "attach_file", "/file-resources", 60)
	ensureLink(menuIDByChildRoute("/users"), "File Resource Types", "/file-resource-types", "folder_open", "/file-resource-types", 50)
}

func menuIDByName(name string) *uint {
	var row menu.Menu
	err := database.GORM_DB.Where("name = ? AND type = ?", name, menu.TypeGroup).First(&row).Error
	if err != nil {
		return nil
	}
	return &row.ID
}

func menuIDByChildRoute(route string) *uint {
	var row menu.Menu
	err := database.GORM_DB.Where("route = ? AND parent_id IS NOT NULL", route).First(&row).Error
	if err != nil || row.ParentID == nil {
		return nil
	}
	return row.ParentID
}

func ensureLink(parentID *uint, name, route, icon, apiPath string, sort int) {
	if parentID == nil {
		log.Printf("[seeder] skip menu %q: no parent\n", name)
		return
	}
	var existing int64
	if err := database.GORM_DB.Model(&menu.Menu{}).Where("route = ?", route).Count(&existing).Error; err != nil {
		log.Printf("[seeder] menu lookup %q: %v\n", name, err)
		return
	}
	if existing > 0 {
		return
	}
	code := authoritycode.DeriveCode(http.MethodGet, "/api/v1"+apiPath)
	var auth authority.Authority
	if err := database.GORM_DB.Where("code = ?", code).First(&auth).Error; err != nil {
		log.Printf("[seeder] skip menu %q: authority %s: %v\n", name, code, err)
		return
	}
	record := menu.Menu{
		Name:        name,
		Type:        menu.TypeLink,
		Route:       &route,
		Icon:        &icon,
		AuthorityID: auth.ID,
		ParentID:    parentID,
		SortOrder:   sort,
		Active:      true,
	}
	if err := database.GORM_DB.Create(&record).Error; err != nil {
		log.Printf("[seeder] could not create menu %q: %v\n", name, err)
	}
}

func ensureMenu(item menuSeed, parentID *uint) int {
	var existing menu.Menu
	q := database.GORM_DB.Where("name = ? AND type = ?", item.Name, item.Type)
	if parentID == nil {
		q = q.Where("parent_id IS NULL")
	} else {
		q = q.Where("parent_id = ?", *parentID)
	}

	err := q.First(&existing).Error
	created := 0
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var auth authority.Authority
		if err := database.GORM_DB.Where("code = ?", item.Authority).First(&auth).Error; err != nil {
			log.Printf("[seeder] skip menu %q: authority %s: %v\n", item.Name, item.Authority, err)
			return 0
		}
		record := menu.Menu{
			Name:        item.Name,
			Type:        item.Type,
			Icon:        &item.Icon,
			AuthorityID: auth.ID,
			ParentID:    parentID,
			SortOrder:   item.SortOrder,
			Active:      true,
		}
		if item.Route != "" {
			record.Route = &item.Route
		}
		if err := database.GORM_DB.Create(&record).Error; err != nil {
			log.Printf("[seeder] could not create menu %q: %v\n", item.Name, err)
			return 0
		}
		existing = record
		created = 1
	} else if err != nil {
		log.Printf("[seeder] menu lookup %q: %v\n", item.Name, err)
		return 0
	}

	for _, child := range item.Children {
		created += ensureMenu(child, &existing.ID)
	}
	return created
}
