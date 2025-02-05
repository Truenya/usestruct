package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestFuncDeclToKey(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "simple function",
			code:     "package test\nfunc Test() {}",
			expected: "Test",
		},
		{
			name:     "method with pointer receiver",
			code:     "package test\ntype Test struct{}\nfunc (t *Test) Test() {}",
			expected: "Test.Test",
		},
		{
			name:     "method with value receiver",
			code:     "package test\ntype Test struct{}\nfunc (t Test) Test() {}",
			expected: "Test.Test",
		},
		{
			name:     "nil function",
			code:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var funcDecl *ast.FuncDecl
			if tt.code != "" {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, "", tt.code, 0)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}
				funcDecl = file.Decls[len(file.Decls)-1].(*ast.FuncDecl)
			}

			mocker := NewMocker()
			if funcDecl != nil && funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				testType := types.NewNamed(
					types.NewTypeName(token.NoPos, nil, "Test", nil),
					types.NewStruct(nil, nil),
					nil,
				)

				if star, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
					mocker.AddType(funcDecl.Recv.List[0].Type, types.NewPointer(testType))
					mocker.AddType(star.X, testType)
				} else {
					mocker.AddType(funcDecl.Recv.List[0].Type, testType)
				}
			}

			analyzer := mocker.Analyzer()
			got := analyzer.funcDeclToKey(funcDecl)
			if got != tt.expected {
				t.Errorf("funcDeclToKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}
