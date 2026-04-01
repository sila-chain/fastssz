package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"unsafe"
)

const bytesPerLengthOffset = 4

type unmarshallUints interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// UnmarshallUint unmarshals a little endian uint8, uint16, uint32, or uint64 from the src input.
func UnmarshallUint[T unmarshallUints](src []byte) T {
	var zero T
	switch unsafe.Sizeof(zero) {
	case 1:
		return T(src[0])
	case 2:
		return T(binary.LittleEndian.Uint16(src[:2]))
	case 4:
		return T(binary.LittleEndian.Uint32(src[:4]))
	case 8:
		return T(binary.LittleEndian.Uint64(src[:8]))
	default:
		panic("unsupported uint size")
	}
}

// UnmarshallUint64 unmarshals a little endian uint64 from the src input.
//
// Deprecated: use UnmarshallUint instead.
func UnmarshallUint64(src []byte) uint64 {
	return UnmarshallUint[uint64](src)
}

// UnmarshallUint32 unmarshals a little endian uint32 from the src input.
//
// Deprecated: use UnmarshallUint instead.
func UnmarshallUint32(src []byte) uint32 {
	return UnmarshallUint[uint32](src)
}

// UnmarshallUint16 unmarshals a little endian uint16 from the src input.
//
// Deprecated: use UnmarshallUint instead.
func UnmarshallUint16(src []byte) uint16 {
	return UnmarshallUint[uint16](src)
}

// UnmarshallUint8 unmarshals a little endian uint8 from the src input.
//
// Deprecated: use UnmarshallUint instead.
func UnmarshallUint8(src []byte) uint8 {
	return UnmarshallUint[uint8](src)
}

// UnmarshalBool unmarshals a boolean from the src input
func UnmarshalBool(src []byte) bool {
	if src[0] == 1 {
		return true
	}
	return false
}

var (
	ErrOffsetExceedsSize           = errors.New("offset exceeds size of buffer")
	ErrOffsetOrdering              = errors.New("offset is less than previous offset")
	ErrDynamicLengthTooShort       = errors.New("buffer too small to hold an offset")
	ErrDynamicLengthNotOffsetSized = errors.New("list offsets must be multiples of the offset size (4)")
	ErrDynamicLengthExceedsMax     = errors.New("list length longer than ssz max length for the type")
	ErrInvalidEncoding             = errors.New("invalid encoding")
)

// ValidateBitlist validates that the bitlist is correct
func ValidateBitlist(buf []byte, bitLimit uint64) error {
	byteLen := len(buf)
	if byteLen == 0 {
		return fmt.Errorf("bitlist empty, it does not have length bit")
	}
	// Maximum possible bytes in a bitlist with provided bitlimit.
	maxBytes := (bitLimit >> 3) + 1
	if byteLen > int(maxBytes) {
		return fmt.Errorf("unexpected number of bytes, got %d but found %d", byteLen, maxBytes)
	}

	// The most significant bit is present in the last byte in the array.
	last := buf[byteLen-1]
	if last == 0 {
		return fmt.Errorf("trailing byte is zero")
	}

	// Determine the position of the most significant bit.
	msb := bits.Len8(last)

	// The absolute position of the most significant bit will be the number of
	// bits in the preceding bytes plus the position of the most significant
	// bit. Subtract this value by 1 to determine the length of the bitlist.
	numOfBits := uint64(8*(byteLen-1) + msb - 1)

	if numOfBits > bitLimit {
		return fmt.Errorf("too many bits")
	}
	return nil
}

// DecodeDynamicLength decodes the length from the dynamic input
func DecodeDynamicLength(buf []byte, maxSize int) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if len(buf) < 4 {
		return 0, ErrDynamicLengthTooShort
	}
	o := int(binary.LittleEndian.Uint32(buf))
	if o%bytesPerLengthOffset != 0 || o == 0 {
		return 0, ErrDynamicLengthNotOffsetSized
	}

	length := o / bytesPerLengthOffset
	if length > maxSize {
		return 0, ErrDynamicLengthExceedsMax
	}

	return length, nil
}

// UnmarshalDynamic unmarshals the dynamic items from the input
func UnmarshalDynamic(src []byte, length int, f func(indx int, b []byte) error) error {
	var err error
	size := uint64(len(src))

	if length == 0 {
		if size != 0 && size != 4 {
			return ErrSize
		}
		return nil
	}

	indx := 0
	dst := src

	var offset, endOffset uint64
	offset, dst = ReadOffset(src), dst[4:]

	for {
		if length != 1 {
			endOffset, dst, err = safeReadOffset(dst)
			if err != nil {
				return err
			}
		} else {
			endOffset = uint64(len(src))
		}
		if offset > endOffset {
			return ErrOffsetOrdering
		}
		if endOffset > size {
			return ErrOffsetExceedsSize
		}

		err := f(indx, src[offset:endOffset])
		if err != nil {
			return err
		}

		indx++

		offset = endOffset
		if length == 1 {
			break
		}
		length--
	}
	return nil
}

func DivideInt2(a, b, max int) (int, error) {
	num, ok := DivideInt(a, b)
	if !ok {
		return 0, fmt.Errorf("a is not evenly divisble by b")
	}
	if num > max {
		return 0, fmt.Errorf("a/b is greater than max")
	}
	return num, nil
}

// DivideInt divides the int fully
func DivideInt(a, b int) (int, bool) {
	return a / b, a%b == 0
}

type uints interface {
	~uint8 | ~uint16 | ~uint64
}

// ExtendUint extends an unsigned integer buffer to a given size.
func ExtendUint[T uints](b []T, needLen int) []T {
	b = b[:cap(b)]
	if n := needLen - cap(b); n > 0 {
		b = append(b, make([]T, n)...)
	}
	return b[:needLen]
}

// ExtendUint64 extends a uint64 buffer to a given size.
//
// Deprecated: use ExtendUint instead.
func ExtendUint64(b []uint64, needLen int) []uint64 {
	return ExtendUint(b, needLen)
}

// ExtendUint16 extends a uint16 buffer to a given size.
//
// Deprecated: use ExtendUint instead.
func ExtendUint16(b []uint16, needLen int) []uint16 {
	return ExtendUint(b, needLen)
}

// ExtendUint8 extends a uint8 buffer to a given size.
//
// Deprecated: use ExtendUint instead.
func ExtendUint8(b []uint8, needLen int) []uint8 {
	return ExtendUint(b, needLen)
}

// ReadOffset reads an offset from buf
func ReadOffset(buf []byte) uint64 {
	return uint64(binary.LittleEndian.Uint32(buf))
}

func safeReadOffset(buf []byte) (uint64, []byte, error) {
	if len(buf) < 4 {
		return 0, nil, fmt.Errorf("")
	}
	offset := ReadOffset(buf)
	return offset, buf[4:], nil
}

func DecodeBool(src []byte) (bool, error) {
	if src[0] == 1 {
		return true, nil
	}
	if src[0] == 0 {
		return false, nil
	}
	return false, ErrInvalidEncoding
}
