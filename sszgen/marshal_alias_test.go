package main

import "testing"

func TestMarshalUsesGenericMarshalUintForUintAlias(t *testing.T) {
	v := &Value{
		name: "ValidatorIndex",
		t:    TypeUint,
		s:    8,
		obj:  "ValidatorIndex",
	}

	got := v.marshal()
	want := "dst = ssz.MarshalUint(dst, ::.ValidatorIndex)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
