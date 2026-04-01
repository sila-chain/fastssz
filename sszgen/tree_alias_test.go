package main

import "testing"

func TestGetTreeUsesGenericAddUintForUintAlias(t *testing.T) {
	v := &Value{
		name: "ValidatorIndex",
		t:    TypeUint,
		s:    8,
		obj:  "ValidatorIndex",
	}

	got := v.getTree()
	want := "ssz.AddUint(w, ::.ValidatorIndex)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
