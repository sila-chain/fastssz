package main

import (
	"strings"
	"testing"
)

func TestHashRootsUsesGenericAppendUintForUintAliasElements(t *testing.T) {
	v := &Value{
		name: "ValidatorIndices",
		t:    TypeVector,
		s:    512,
		e: &Value{
			t:   TypeUint,
			s:   8,
			obj: "ValidatorIndex",
		},
	}

	got := v.hashRoots(false, v.e.t)
	if !strings.Contains(got, "ssz.AppendUint(hh, i)") {
		t.Fatalf("expected generic uint append in generated hash roots, got:\n%s", got)
	}
	if strings.Contains(got, "hh.AppendUint64(") {
		t.Fatalf("unexpected typed uint append in generated hash roots:\n%s", got)
	}
}

func TestHashTreeRootUsesGenericPutUintForUintAlias(t *testing.T) {
	v := &Value{
		name: "ValidatorIndex",
		t:    TypeUint,
		s:    8,
		obj:  "ValidatorIndex",
	}

	got := v.hashTreeRoot("value", false)
	if !strings.Contains(got, "ssz.PutUint(hh, value)") {
		t.Fatalf("expected generic uint put in generated hash root, got:\n%s", got)
	}
	if strings.Contains(got, "hh.PutUint64(") {
		t.Fatalf("unexpected typed uint put in generated hash root:\n%s", got)
	}
}
