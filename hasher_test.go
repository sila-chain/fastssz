package ssz

import (
	"bytes"
	"testing"
)

type validatorIndex uint64

func TestNextPowerOfTwo(t *testing.T) {
	cases := []struct {
		Num, Res uint64
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{10, 16},
		{11, 16},
		{13, 16},
	}
	for _, c := range cases {
		if next := nextPowerOfTwo(c.Num); uint64(next) != c.Res {
			t.Fatalf("num %d, expected %d but found %d", c.Num, c.Res, next)
		}
	}
}

func TestMerkleize8ByteVector(t *testing.T) {
	result := merkleizeInput([]byte{'1', '2', '3', '4', '5', '6', '7', '8'}, 0)
	if !bytes.Equal(result, []byte{49, 50, 51, 52, 53, 54, 55, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatalf("Unexpected result: %v", result)
	}
}

func TestAppendUintAcceptsNamedUint64(t *testing.T) {
	var h Hasher
	AppendUint(&h, validatorIndex(7))

	want := MarshalUint64(nil, 7)
	if !bytes.Equal(h.buf, want) {
		t.Fatalf("unexpected result: %v", h.buf)
	}
}
