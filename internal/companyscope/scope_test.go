package companyscope

import "testing"

func TestScopeAllows(t *testing.T) {
	open := Scope{Admin: true}
	if !open.Allows(9) {
		t.Fatal("administrator should manage any company")
	}
	limited := Scope{OperatorIDs: []uint{3, 5}}
	if !limited.Allows(5) {
		t.Fatal("member company should be allowed")
	}
	if limited.Allows(9) {
		t.Fatal("another company should be refused")
	}
}
