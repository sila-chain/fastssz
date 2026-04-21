package spectests

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

// specByteListRoot computes hash_tree_root(ByteList[N](data)) per the SSZ spec
// using only crypto/sha256. It is an independent ground truth for comparing
// against the sszgen-generated HashTreeRoot implementation.
func specByteListRoot(data []byte, maxBytes uint64) [32]byte {
	padded := append([]byte(nil), data...)
	if r := len(padded) % 32; r != 0 {
		padded = append(padded, make([]byte, 32-r)...)
	}
	limit := (maxBytes + 31) / 32
	treeSize := uint64(1)
	for treeSize < limit {
		treeSize *= 2
	}
	layer := make([][32]byte, treeSize)
	for i := 0; i*32 < len(padded); i++ {
		copy(layer[i][:], padded[i*32:(i+1)*32])
	}
	var buf [64]byte
	for len(layer) > 1 {
		next := make([][32]byte, len(layer)/2)
		for i := range next {
			copy(buf[:32], layer[2*i][:])
			copy(buf[32:], layer[2*i+1][:])
			next[i] = sha256.Sum256(buf[:])
		}
		layer = next
	}
	var lenBytes [32]byte
	binary.LittleEndian.PutUint64(lenBytes[:8], uint64(len(data)))
	copy(buf[:32], layer[0][:])
	copy(buf[32:], lenBytes[:])
	return sha256.Sum256(buf[:])
}

func TestByteListContainerHashTreeRootMatchesSpec(t *testing.T) {
	const maxBytes uint64 = 256

	cases := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"one_byte", []byte{0xAA}},
		{"half_chunk", bytes.Repeat([]byte{0xAB}, 16)},
		{"one_chunk", bytes.Repeat([]byte{0xCD}, 32)},
		// 33 bytes spans two chunks — the input size that regresses the
		// PutBytes-then-MerkleizeWithMixin double-hash bug.
		{"two_chunks", bytes.Repeat([]byte{0xEF}, 33)},
		{"many_chunks", bytes.Repeat([]byte{0x11}, 200)},
		{"full_capacity", bytes.Repeat([]byte{0x22}, 256)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &ByteListContainer{Data: tc.data}
			got, err := obj.HashTreeRoot()
			if err != nil {
				t.Fatal(err)
			}
			// ByteListContainer has one field, so the container root equals
			// the field root (merkleize of 1 leaf is identity).
			want := specByteListRoot(tc.data, maxBytes)
			if got != want {
				t.Fatalf("root mismatch for %s (len=%d)\n want %x\n  got %x",
					tc.name, len(tc.data), want, got)
			}
		})
	}
}
