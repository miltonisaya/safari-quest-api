package user

import (
	"errors"
	"strings"
	"testing"

	"safari-quest-api/config"

	"github.com/google/uuid"
)

func TestParseListFilter(t *testing.T) {
	id := uuid.New()
	filter, err := parseListFilter(id.String(), "true")
	if err != nil {
		t.Fatal(err)
	}
	if filter.RoleUUID == nil || *filter.RoleUUID != id || filter.Active == nil || !*filter.Active {
		t.Fatalf("filter = %+v", filter)
	}

	filter, err = parseListFilter("  ", " false ")
	if err != nil || filter.RoleUUID != nil || filter.Active == nil || *filter.Active {
		t.Fatalf("blank role should be ignored, got %+v %v", filter, err)
	}

	if _, err = parseListFilter("not-a-uuid", ""); !errors.Is(err, ErrInvalidRoleFilter) {
		t.Fatalf("role error = %v", err)
	}
	if _, err = parseListFilter("", "yes"); !errors.Is(err, ErrInvalidStatusFilter) {
		t.Fatalf("status error = %v", err)
	}
}

func TestCompanyUserScope(t *testing.T) {
	own := uuid.New()
	other := uuid.New()
	caller := []uuid.UUID{own}

	if err := mayManage(true, false, nil, []uuid.UUID{other}); err != nil {
		t.Fatalf("administrator = %v", err)
	}
	if err := mayManage(false, false, caller, caller); !errors.Is(err, ErrOutsideCompany) {
		t.Fatalf("other company = %v", err)
	}
	if err := mayManage(false, true, caller, []uuid.UUID{other}); !errors.Is(err, ErrOutsideCompany) {
		t.Fatalf("role outside the company = %v", err)
	}
	if err := mayManage(false, true, caller, caller); err != nil {
		t.Fatalf("colleague = %v", err)
	}
	if err := mayGrant(false, caller, []uuid.UUID{other}); !errors.Is(err, ErrRoleNotAllowed) {
		t.Fatalf("grant = %v", err)
	}
	if err := mayGrant(false, caller, caller); err != nil {
		t.Fatalf("grant own role = %v", err)
	}
}

func TestGuardOwnAccount(t *testing.T) {
	self := uuid.New()
	other := uuid.New()
	off := false
	on := true

	if err := guardOwnAccount(self, self, &off, false); !errors.Is(err, ErrCannotDeactivateSelf) {
		t.Fatalf("deactivate self = %v", err)
	}
	if err := guardOwnAccount(self, self, nil, true); !errors.Is(err, ErrCannotDeleteSelf) {
		t.Fatalf("delete self = %v", err)
	}
	if err := guardOwnAccount(self, self, &on, false); err != nil {
		t.Fatalf("activate self = %v", err)
	}
	if err := guardOwnAccount(self, self, nil, false); err != nil {
		t.Fatalf("reset self = %v", err)
	}
	if err := guardOwnAccount(self, other, &off, true); err != nil {
		t.Fatalf("other account = %v", err)
	}
}

func TestDefaultPasswordUnset(t *testing.T) {
	prev := config.App.AdminDefaultPassword
	t.Cleanup(func() { config.App.AdminDefaultPassword = prev })

	config.App.AdminDefaultPassword = "  "
	if _, err := defaultPasswordHash(); !errors.Is(err, ErrDefaultPasswordUnset) {
		t.Fatalf("blank password error = %v", err)
	}

	config.App.AdminDefaultPassword = "test-only-secret"
	hashed, err := defaultPasswordHash()
	if err != nil {
		t.Fatal(err)
	}
	if hashed == "" || hashed == "test-only-secret" || strings.Contains(hashed, "test-only-secret") {
		t.Fatal("hash must not be the raw password")
	}
}
