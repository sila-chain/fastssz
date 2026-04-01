package main

import "testing"

func TestUnmarshalUsesGenericUnmarshallUintForUintAlias(t *testing.T) {
	v := &Value{
		name: "ValidatorIndex",
		t:    TypeUint,
		s:    8,
		obj:  "ValidatorIndex",
	}

	got := v.unmarshal("buf")
	want := "::.ValidatorIndex = ssz.UnmarshallUint[ValidatorIndex](buf)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
