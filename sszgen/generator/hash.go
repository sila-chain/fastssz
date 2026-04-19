package generator

import (
	"fmt"
	"strings"
)

func (e *env) hashTreeRoot(name string, v *Value) string {
	tmpl := `// HashTreeRoot ssz hashes the {{.name}} object
	func (:: *{{.name}}) HashTreeRoot() ([32]byte, error) {
		return ssz.HashWithDefaultHasher(::)
	}
	
	// HashTreeRootWith ssz hashes the {{.name}} object with a hasher	
	func (:: *{{.name}}) HashTreeRootWith(hh *ssz.Hasher) (err error) {
		{{.hashTreeRoot}}
		return
	}`

	data := map[string]interface{}{
		"name":         name,
		"hashTreeRoot": v.hashTreeRootContainer(true),
	}
	str := execTmpl(tmpl, data)
	return appendObjSignature(str, v)
}

func (v *Value) hashRoots(isList bool, elem Type) string {
	subName := "i"
	if v.e.c {
		subName += "[:]"
	}
	inner := ""
	if !v.e.c && elem == TypeBytes {
		inner = fmt.Sprintf(`if len(i) != %d {
			err = ssz.ErrBytesLength
			return
		}
		`, v.e.s)
	}

	appendExpr := ""
	appendFn := ""
	var elemSize uint64
	if elem == TypeBytes {
		if v.e.s != 32 {
			appendFn = "PutBytes"
			elemSize = v.e.s
		} else {
			appendFn = "Append"
			elemSize = 32
		}
	} else {
		appendExpr = fmt.Sprintf("ssz.AppendUint(hh, %s)", subName)
		elemSize = uint64(v.e.fixedSize())
	}
	if appendExpr == "" {
		appendExpr = fmt.Sprintf("hh.%s(%s)", appendFn, subName)
	}

	merkleize := "hh.Merkleize(subIndx)"
	if isList {
		isComplex := v.e.t == TypeBytes
		tmpl := `
		numItems := uint64(len(::.{{.name}}))
		hh.MerkleizeWithMixin(subIndx, numItems, {{if .isComplex}} {{.listSize}} {{ else }} ssz.CalculateLimit({{.listSize}}, numItems, {{.elemSize}}) {{ end }})`
		merkleize = execTmpl(tmpl, map[string]interface{}{
			"name":      v.name,
			"listSize":  v.s,
			"elemSize":  elemSize,
			"isComplex": isComplex,
		})
		if elem == TypeUint {
			merkleize = "hh.FillUpTo32()\n" + merkleize
		}
	}

	tmpl := `{
		{{.outer}}subIndx := hh.Index()
		for _, i := range ::.{{.name}} {
			{{.inner}}{{.appendExpr}}
		}
		{{.merkleize}}
	}`
	return execTmpl(tmpl, map[string]interface{}{
		"outer":      v.validate(),
		"inner":      inner,
		"name":       v.name,
		"appendExpr": appendExpr,
		"merkleize":  merkleize,
	})
}

func (v *Value) hashTreeRoot(name string, appendBytes bool) string {
	if name == "" {
		name = "::." + v.name
	}
	switch v.t {
	case TypeContainer, TypeReference:
		return v.hashTreeRootContainer(false)
	case TypeBytes:
		if v.c {
			name += "[:]"
		}
		if v.isFixed() {
			tmpl := `{{.validate}}hh.PutBytes({{.name}})`
			return execTmpl(tmpl, map[string]interface{}{
				"validate": v.validate(),
				"name":     name,
			})
		}
		// PutBytes auto-merkleizes when len > 32, which double-hashes against
		// MerkleizeWithMixin below and yields the wrong root.
		_ = appendBytes
		tmpl := `{
	elemIndx := hh.Index()
	byteLen := uint64(len({{.name}}))
	if byteLen > {{.maxLen}} {
		err = ssz.ErrIncorrectListSize
		return
    }
	hh.AppendBytes32({{.name}})
	hh.MerkleizeWithMixin(elemIndx, byteLen, ({{.maxLen}}+31)/32)
}`
		return execTmpl(tmpl, map[string]interface{}{
			"name":   name,
			"maxLen": v.m,
		})
	case TypeUint:
		return fmt.Sprintf("ssz.PutUint(hh, %s)", name)
	case TypeBitList:
		tmpl := `if len({{.name}}) == 0 {
			err = ssz.ErrEmptyBitlist
			return
		}
		hh.PutBitlist({{.name}}, {{.size}})
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": name,
			"size": v.m,
		})
	case TypeBool:
		return fmt.Sprintf("hh.PutBool(%s)", name)
	case TypeVector:
		if v.e.t == TypeContainer {
			tmpl := `{
				subIndx := hh.Index()
				for _, elem := range {{.name}} {
					if err = elem.HashTreeRootWith(hh); err != nil {
						return
					}
				}
				hh.Merkleize(subIndx)
			}`
			return execTmpl(tmpl, map[string]interface{}{
				"name": name,
			})
		}
		return v.hashRoots(false, v.e.t)
	case TypeList:
		if v.e.isFixed() && (v.e.t == TypeUint || v.e.t == TypeBytes) {
			return v.hashRoots(true, v.e.t)
		}
		tmpl := `{
			subIndx := hh.Index()
			num := uint64(len({{.name}}))
			if num > {{.num}} {
				err = ssz.ErrIncorrectListSize
				return
			}
			for _, elem := range {{.name}} {
{{.htrCall}}
			}
			hh.MerkleizeWithMixin(subIndx, num, {{.num}})
		}`
		htrCall := ""
		if v.e.t == TypeBytes {
			htrCall = v.e.hashTreeRoot("elem", true)
		} else {
			htrCall = execTmpl(`if err = elem.HashTreeRootWith(hh); err != nil {
	return
}`,
				map[string]interface{}{"name": name})
		}
		return execTmpl(tmpl, map[string]interface{}{
			"name":    name,
			"num":     v.m,
			"htrCall": htrCall,
		})
	default:
		panic(fmt.Errorf("hash not implemented for type %s", v.t.String()))
	}
}

func (v *Value) hashTreeRootContainer(start bool) string {
	if !start {
		return fmt.Sprintf("if err = ::.%s.HashTreeRootWith(hh); err != nil {\n return\n}", v.name)
	}

	out := []string{}
	for indx, i := range v.o {
		out = append(out, fmt.Sprintf("// Field (%d) '%s'\n%s\n", indx, i.name, i.hashTreeRoot("", false)))
	}

	tmpl := `indx := hh.Index()

	{{.fields}}
	
	hh.Merkleize(indx)`

	return execTmpl(tmpl, map[string]interface{}{
		"fields": strings.Join(out, "\n"),
	})
}
