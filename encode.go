package ssz

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

// MarshalSSZ marshals an object
func MarshalSSZ(m Marshaler) ([]byte, error) {
	buf := make([]byte, m.SizeSSZ())
	return m.MarshalSSZTo(buf[:0])
}

var (
	ErrOffset                = fmt.Errorf("incorrect offset")
	ErrSize                  = fmt.Errorf("incorrect size")
	ErrBytesLength           = fmt.Errorf("bytes array does not have the correct length")
	ErrVectorLength          = fmt.Errorf("vector does not have the correct length")
	ErrListTooBig            = fmt.Errorf("list length is higher than max value")
	ErrEmptyBitlist          = fmt.Errorf("bitlist is empty")
	ErrInvalidVariableOffset = fmt.Errorf("invalid ssz encoding. first variable element offset indexes into fixed value data")
)

func ErrBytesLengthFn(name string, found, expected int) error {
	return fmt.Errorf("%s (%v): expected %d and %d found", name, ErrBytesLength, expected, found)
}

func ErrVectorLengthFn(name string, found, expected int) error {
	return fmt.Errorf("%s (%v): expected %d and %d found", name, ErrBytesLength, expected, found)
}

func ErrListTooBigFn(name string, found, max int) error {
	return fmt.Errorf("%s (%v): max expected %d and %d found", name, ErrListTooBig, max, found)
}

type marshalUints interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// MarshalUint marshals a little endian uint8, uint16, uint32, or uint64 to dst.
func MarshalUint[T marshalUints](dst []byte, i T) []byte {
	var buf [8]byte
	switch unsafe.Sizeof(i) {
	case 1:
		return append(dst, byte(i))
	case 2:
		binary.LittleEndian.PutUint16(buf[:2], uint16(i))
		return append(dst, buf[:2]...)
	case 4:
		binary.LittleEndian.PutUint32(buf[:4], uint32(i))
		return append(dst, buf[:4]...)
	case 8:
		binary.LittleEndian.PutUint64(buf[:8], uint64(i))
		return append(dst, buf[:8]...)
	default:
		panic("unsupported uint size")
	}
}

// MarshalUint64 marshals a little endian uint64 to dst.
//
// Deprecated: use MarshalUint instead.
func MarshalUint64(dst []byte, i uint64) []byte {
	return MarshalUint(dst, i)
}

// MarshalUint32 marshals a little endian uint32 to dst.
//
// Deprecated: use MarshalUint instead.
func MarshalUint32(dst []byte, i uint32) []byte {
	return MarshalUint(dst, i)
}

// MarshalUint16 marshals a little endian uint16 to dst.
//
// Deprecated: use MarshalUint instead.
func MarshalUint16(dst []byte, i uint16) []byte {
	return MarshalUint(dst, i)
}

// MarshalUint8 marshals a little endian uint8 to dst.
//
// Deprecated: use MarshalUint instead.
func MarshalUint8(dst []byte, i uint8) []byte {
	return MarshalUint(dst, i)
}

// MarshalBool marshals a boolean to dst
func MarshalBool(dst []byte, b bool) []byte {
	if b {
		return append(dst, 1)
	}
	return append(dst, 0)
}

// WriteOffset writes an offset to dst
func WriteOffset(dst []byte, i int) []byte {
	return MarshalUint(dst, uint32(i))
}
