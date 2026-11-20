// Package companyscope decides which tour companies the signed-in account
// may manage. Administrators are unrestricted. Everyone else is limited to
// the companies they belong to.
package companyscope

import (
	"errors"

	"safari-quest-api/internal/role"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
)

// ErrNoCompany is returned when the account is not an administrator and is
// not linked to any tour company.
var ErrNoCompany = errors.New("You can only manage records in your company")

const administratorRoleCode = "ADMINISTRATOR"

// Scope is the caller's company boundary for one request.
type Scope struct {
	Admin       bool
	OperatorIDs []uint
}

// Load reads the signed-in user and their companies.
func Load(id uuid.UUID) (Scope, error) {
	account, err := user.FindByUUID(id)
	if err != nil {
		return Scope{}, err
	}
	if isAdministrator(account.Roles) {
		return Scope{Admin: true}, nil
	}
	ids, err := user.OperatorIDsForUser(account.ID)
	if err != nil {
		return Scope{}, err
	}
	if len(ids) == 0 {
		return Scope{}, ErrNoCompany
	}
	return Scope{OperatorIDs: ids}, nil
}

// Allows reports whether this scope may act on the company.
func (scope Scope) Allows(operatorID uint) bool {
	if scope.Admin {
		return true
	}
	for _, id := range scope.OperatorIDs {
		if id == operatorID {
			return true
		}
	}
	return false
}

func isAdministrator(roles []role.Role) bool {
	for _, item := range roles {
		if item.Code == administratorRoleCode {
			return true
		}
	}
	return false
}
