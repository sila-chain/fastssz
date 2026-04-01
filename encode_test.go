package ssz

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

type slot uint64

func TestMarshalUintAcceptsNamedUint64(t *testing.T) {
	got := MarshalUint(nil, slot(9))
	want := []byte{9, 0, 0, 0, 0, 0, 0, 0}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestDeprecatedMarshalWrappersCallMarshalUint(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "encode.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"MarshalUint64", "MarshalUint32", "MarshalUint16", "MarshalUint8"} {
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
			ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) != 1 {
				t.Fatalf("expected %s to return a direct call", name)
			}
			call, ok := ret.Results[0].(*ast.CallExpr)
			if !ok {
				t.Fatalf("expected %s to return a call", name)
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok || ident.Name != "MarshalUint" {
				t.Fatalf("expected %s to call MarshalUint", name)
			}
		}
		if !found {
			t.Fatalf("did not find %s", name)
		}
	}
}

func TestWriteOffsetUsesMarshalUint(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "encode.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "WriteOffset" {
			continue
		}
		ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			t.Fatal("expected WriteOffset to return a direct call")
		}
		call, ok := ret.Results[0].(*ast.CallExpr)
		if !ok {
			t.Fatal("expected WriteOffset to return a call")
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "MarshalUint" {
			t.Fatal("expected WriteOffset to call MarshalUint")
		}
		return
	}
	t.Fatal("did not find WriteOffset")
}
