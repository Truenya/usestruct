package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestInitArgsMap(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected map[string]int
	}{
		{
			name:     "separate parameters",
			code:     "package test\nfunc foo(a int, b string, c float64) {}",
			expected: map[string]int{"int": 1, "string": 1, "float64": 1},
		},
		{
			name:     "grouped parameters",
			code:     "package test\nfunc foo(a, b, c int) {}",
			expected: map[string]int{"int": 3},
		},
		{
			name:     "mixed parameters",
			code:     "package test\nfunc foo(a, b int, c string, d, e float64) {}",
			expected: map[string]int{"int": 2, "string": 1, "float64": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("failed to parse code: %v", err)
			}
			funcDecl := file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
			params := funcDecl.Type.Params.List
			mocker := NewMocker()
			args := mocker.Analyzer().initArgsMap(params)
			if len(args) != len(tt.expected) {
				t.Errorf("expected %d unique types, got %d", len(tt.expected), len(args))
			}
			for typ, count := range tt.expected {
				if args[typ] != count {
					t.Errorf("expected %d of type %s, got %d", count, typ, args[typ])
				}
			}
		})
	}
}

// func TestModifyArgsByLower(t *testing.T) {
// 	fset := token.NewFileSet()
// 	code := "package test\nfunc lower(a, b int) {}\nfunc upper(a, b, c int) {}"
// 	file, err := parser.ParseFile(fset, "", code, 0)
// 	if err != nil {
// 		t.Fatalf("failed to parse code: %v", err)
// 	}
// 	lower := file.Decls[0].(*ast.FuncDecl)
// 	upper := file.Decls[1].(*ast.FuncDecl)
// 	args := initArgsMap(upper.Type.Params.List)

// 	mocker := NewMocker()
// 	analyzer := mocker.Analyzer()
// 	inter := analyzer.modifyArgsByLower(args, lower)
// 	if len(inter) != 1 {
// 		t.Errorf("expected 1 type in intersection, got %d", len(inter))
// 	}
// 	if inter["int"] != 2 {
// 		t.Errorf("expected 2 of type int in intersection, got %d", inter["int"])
// 	}
// }
