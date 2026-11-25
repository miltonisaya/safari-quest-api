package dbtypes

import (
	"reflect"
	"testing"
)

func TestStringListRoundTrip(t *testing.T) {
	in := StringList{"Park fees", "Game drives"}
	v, err := in.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out StringList
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip changed the list: %v -> %v", in, out)
	}
}

func TestStringListNilIsEmptyArray(t *testing.T) {
	var l StringList
	v, _ := l.Value()
	if v != "[]" {
		t.Fatalf("nil list stored as %v, want []", v)
	}
	var out StringList
	if err := out.Scan([]byte("null")); err != nil {
		t.Fatal(err)
	}
	if out == nil || len(out) != 0 {
		t.Fatalf("null column scanned as %#v, want empty list", out)
	}
}

func TestStringListClean(t *testing.T) {
	got := StringList{"  Park fees ", "", "\t", "Meals\n"}.Clean()
	want := StringList{"Park fees", "Meals"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Clean() = %v, want %v", got, want)
	}
}
