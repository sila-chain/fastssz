package generator

func (v *Value) validate() string {
	switch v.t {
	case TypeBitList, TypeBytes:
		if v.c {
			return ""
		}
		cmp := "!="
		if !v.isFixed() {
			cmp = ">"
		}

		tmpl := `if size := len(::.{{.name}}); size {{.cmp}} {{.size}} {
			err = ssz.ErrBytesLengthFn("--.{{.name}}", size, {{.size}})
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"cmp":  cmp,
			"name": v.name,
			"size": v.s,
		})

	case TypeVector:
		if v.c {
			return ""
		}
		tmpl := `if size := len(::.{{.name}}); size != {{.size}} {
			err = ssz.ErrVectorLengthFn("--.{{.name}}", size, {{.size}})
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": v.s,
		})

	case TypeList:
		tmpl := `if size := len(::.{{.name}}); size > {{.size}} {
			err = ssz.ErrListTooBigFn("--.{{.name}}", size, {{.size}})
			return
		}
		`
		return execTmpl(tmpl, map[string]interface{}{
			"name": v.name,
			"size": v.s,
		})

	default:
		return ""
	}
}
