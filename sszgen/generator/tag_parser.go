package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type tokenState int

const (
	tsBegin tokenState = iota
	tsLabel
	tsValue
	tsCloseTick
)

func GetSSZTags(tag string) (map[string]string, error) {
	var lastErr error
	accumulateError := func(_ *scanner.Scanner, msg string) {
		lastErr = errors.New(msg)
	}

	sr := strings.NewReader(tag)
	sc := scanner.Scanner{}
	sc.Init(sr)
	sc.Filename = "tag"
	sc.Mode ^= scanner.ScanRawStrings
	sc.Error = accumulateError

	var labelStr string
	var state tokenState
	tags := make(map[string]string)
	for tok := sc.Scan(); tok != scanner.EOF; tok = sc.Scan() {
		if lastErr != nil {
			return nil, fmt.Errorf("GetSSZTags failed: token scanner error = %s", lastErr)
		}
		if state == tsCloseTick {
			return nil, errors.New("GetSSZTags failed: undefined behavior when scanning beyond the end of the tag")
		}
		txt := sc.TokenText()
		switch txt {
		case "`":
			if state == tsLabel {
				state = tsCloseTick
				continue
			}
			if state == tsBegin {
				state = tsLabel
				continue
			}
		case ":":
			if state == tsLabel {
				state = tsValue
				continue
			}
		case "\"":
			continue
		default:
			if state == tsValue {
				tags[labelStr] = trimTagQuotes(txt)
				state = tsLabel
				labelStr = ""
				continue
			}
			if state == tsLabel {
				labelStr += txt
				continue
			}
		}
	}
	return tags, nil
}

var nilInt *int

func isBitList(tags map[string]string) bool {
	for k, v := range tags {
		if k == "ssz" {
			parts := strings.Split(v, ",")
			for _, p := range parts {
				if p == "bitlist" {
					return true
				}
			}
		}
	}
	return false
}

func extractSSZDimensions(tag string) ([]*SSZDimension, error) {
	tags, err := GetSSZTags(tag)
	if err != nil {
		return nil, err
	}
	sszSizes, sizeDefined := tags["ssz-size"]
	sszMax, maxDefined := tags["ssz-max"]
	if !sizeDefined && !maxDefined {
		return nil, fmt.Errorf("No ssz-size or ssz-max tags found for element. tag=%s", tag)
	}

	sizeSplit := strings.Split(sszSizes, ",")
	maxSplit := strings.Split(sszMax, ",")
	ndims := len(sizeSplit)
	if len(maxSplit) > len(sizeSplit) {
		ndims = len(maxSplit)
	}
	dims := make([]*SSZDimension, ndims)
	for i := 0; i < ndims; i++ {
		isbl := i == ndims-1 && isBitList(tags)
		var szi, mxi string
		if len(sizeSplit) > i {
			szi = sizeSplit[i]
		}
		if len(maxSplit) > i {
			mxi = maxSplit[i]
		}
		if szi == "?" && mxi == "?" {
			return nil, fmt.Errorf("At dimension %d both ssz-size and ssz-max had a '?' value. For each dimension, either ssz-size or ssz-max must have a value. Ex: 'ssz-size:\"?,32\" ssz-max:\"100\" defines a List with 100 element limit, containing 32 byte fixed-sized vectors. tag=%s", i, tag)
		}
		switch szi {
		case "?", "":
			if mxi == "?" || mxi == "" {
				return nil, fmt.Errorf("no numeric ssz-size or ssz-max tag for value at dimesion %d, tag=%s", i, tag)
			}
			m, err := strconv.Atoi(mxi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-max at dimension %d, tag=%s. err=%s", mxi, i, tag, err)
			}
			dims[i] = &SSZDimension{isBitlist: isbl, ListLength: &m}
		default:
			s, err := strconv.Atoi(szi)
			if err != nil {
				return nil, fmt.Errorf("atoi failed on value %s for ssz-size at dimension %d, tag=%s. err=%s", szi, i, tag, err)
			}
			dims[i] = &SSZDimension{isBitlist: isbl, VectorLength: &s}
		}
	}
	return dims, nil
}

type SSZDimension struct {
	VectorLength *int
	ListLength   *int
	isBitlist    bool
}

func (dim *SSZDimension) IsVector() bool  { return dim.VectorLength != nilInt }
func (dim *SSZDimension) IsList() bool    { return dim.ListLength != nilInt }
func (dim *SSZDimension) IsBitlist() bool { return dim.isBitlist }
func (dim *SSZDimension) ListLen() int    { return *dim.ListLength }
func (dim *SSZDimension) VectorLen() int  { return *dim.VectorLength }

func (dim *SSZDimension) ValueType() Type {
	if dim.IsVector() {
		return TypeVector
	}
	if dim.IsList() {
		return TypeList
	}
	return TypeUndefined
}

func (dim *SSZDimension) ValueLen() uint64 {
	if dim.IsList() {
		return uint64(dim.ListLen())
	}
	if dim.IsVector() {
		return uint64(dim.VectorLen())
	}
	return 0
}

func trimTagQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
