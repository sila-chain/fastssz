package ssz

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

type wrapperSlot uint64

func TestAddUintAcceptsNamedUint64(t *testing.T) {
	var w Wrapper
	AddUint(&w, wrapperSlot(7))

	want := make([]byte, 32)
	want[0] = 7
	if len(w.nodes) != 1 || !bytes.Equal(w.nodes[0].value, want) {
		t.Fatalf("unexpected node value: %v", w.nodes)
	}
}

func TestDeprecatedAddWrappersCallAddUint(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "wrapper.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"AddUint64", "AddUint32", "AddUint16", "AddUint8"} {
		found := false
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != name {
				continue
			}
			found = true
			if len(fn.Body.List) != 1 {
				t.Fatalf("expected %s to have exactly one statement", name)
			}
			exprStmt, ok := fn.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("expected %s to delegate with a direct call", name)
			}
			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				t.Fatalf("expected %s body to be a call", name)
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok || ident.Name != "AddUint" {
				t.Fatalf("expected %s to call AddUint", name)
			}
		}
		if !found {
			t.Fatalf("did not find %s", name)
		}
	}
}
