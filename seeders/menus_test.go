package seeders

import (
	"testing"

	"safari-quest-api/internal/menu"
)

func TestAppMenuTree(t *testing.T) {
	roots := appMenuTree()
	if len(roots) == 0 {
		t.Fatal("menu tree is empty")
	}

	seenRoute := map[string]bool{}
	var walk func([]menuSeed)
	walk = func(items []menuSeed) {
		orders := map[int]string{}
		for _, item := range items {
			if item.Name == "" || item.Icon == "" || item.Authority == "" {
				t.Errorf("%q is missing a name, icon, or authority", item.Name)
			}
			if prev, ok := orders[item.SortOrder]; ok {
				t.Errorf("sort %d is used by both %q and %q", item.SortOrder, prev, item.Name)
			}
			orders[item.SortOrder] = item.Name

			switch item.Type {
			case menu.TypeLink:
				if item.Route == "" || item.Route[0] != '/' {
					t.Errorf("%q has route %q", item.Name, item.Route)
				}
				if seenRoute[item.Route] {
					t.Errorf("route %s is used more than once", item.Route)
				}
				seenRoute[item.Route] = true
				if len(item.Children) > 0 {
					t.Errorf("link %q has children", item.Name)
				}
			case menu.TypeGroup:
				if item.Route != "" {
					t.Errorf("group %q has route %q", item.Name, item.Route)
				}
				if len(item.Children) == 0 {
					t.Errorf("group %q has no items", item.Name)
				}
				walk(item.Children)
			default:
				t.Errorf("%q has type %q", item.Name, item.Type)
			}
		}
	}
	walk(roots)
	assertDepth(t, roots, 1)

	rootsByName := map[string]menuSeed{}
	for _, item := range roots {
		rootsByName[item.Name] = item
	}
	for _, name := range []string{"Dashboard", "Messages", "Settings", "Security & Administration"} {
		if _, ok := rootsByName[name]; !ok {
			t.Errorf("missing top item %q", name)
		}
	}
	location := childNamed(t, rootsByName["Settings"], "Location Management")
	if !seenRouteUnder(location, "/geographies") || !seenRouteUnder(location, "/geography-types") {
		t.Error("Location Management is missing Geographies or Geography Types")
	}
	if !seenRouteUnder(childNamed(t, rootsByName["Settings"], "Attractions Management"), "/attractions") {
		t.Error("Attractions Management is missing Attractions")
	}
	admin := rootsByName["Security & Administration"]
	for _, route := range []string{"/users", "/roles", "/authorities", "/menus"} {
		if !seenRouteUnder(admin, route) {
			t.Errorf("Security & Administration is missing %s", route)
		}
	}

	for _, route := range []string{
		"/dashboard", "/messages",
		"/users", "/roles", "/authorities", "/menus",
		"/geographies", "/geography-types",
		"/attractions", "/attraction-types", "/attraction-categories", "/tour-operators",
		"/packages", "/file-resources", "/file-resource-types",
	} {
		if !seenRoute[route] {
			t.Errorf("missing menu for %s", route)
		}
	}
}

func assertDepth(t *testing.T, items []menuSeed, depth int) {
	t.Helper()
	if depth > 3 {
		t.Fatalf("menu is nested %d levels; the nav renders three", depth)
	}
	for _, item := range items {
		if len(item.Children) > 0 {
			assertDepth(t, item.Children, depth+1)
		}
	}
}

func childNamed(t *testing.T, parent menuSeed, name string) menuSeed {
	t.Helper()
	for _, child := range parent.Children {
		if child.Name == name {
			return child
		}
	}
	t.Fatalf("%q has no child %q", parent.Name, name)
	return menuSeed{}
}

func seenRouteUnder(item menuSeed, route string) bool {
	if item.Route == route {
		return true
	}
	for _, child := range item.Children {
		if seenRouteUnder(child, route) {
			return true
		}
	}
	return false
}
