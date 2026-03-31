package main

import "testing"

func TestCreateSliceUsesGenericExtendUint(t *testing.T) {
	v := &Value{
		name: "Field",
		t:    TypeList,
		e: &Value{
			t: TypeUint,
			s: 8,
		},
	}

	got := v.createSlice(true)
	want := "::.Field = ssz.ExtendUint(::.Field, num)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
