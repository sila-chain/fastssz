package main

import (
	"strings"
	"testing"
)

func TestHashTreeRootByteListUsesAppendBytes32(t *testing.T) {
	v := &Value{name: "BlockAccessList", t: TypeBytes, m: 1073741824}

	got := v.hashTreeRoot("")
	if !strings.Contains(got, "hh.AppendBytes32(") {
		t.Fatalf("missing AppendBytes32:\n%s", got)
	}
	if strings.Contains(got, "hh.PutBytes(") {
		t.Fatalf("PutBytes double-merkleizes for len > 32:\n%s", got)
	}
}

func TestHashTreeRootFixedBytesUsesPutBytes(t *testing.T) {
	v := &Value{name: "ParentHash", t: TypeBytes, s: 32, fixed: true}

	got := v.hashTreeRoot("")
	if !strings.Contains(got, "hh.PutBytes(") {
		t.Fatalf("fixed Bytes32 should use PutBytes:\n%s", got)
	}
	if strings.Contains(got, "MerkleizeWithMixin") {
		t.Fatalf("fixed Bytes32 should not mix in length:\n%s", got)
	}
}
