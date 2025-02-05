package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
)

// createTestFuncDecl создает тестовую функцию с заданным именем и количеством параметров типа int
func createTestFuncDecl(name string, paramCount int) *ast.FuncDecl {
	params := make([]*ast.Field, paramCount)
	for i := range paramCount {
		params[i] = &ast.Field{
			Names: []*ast.Ident{{Name: fmt.Sprintf("a%d", i)}},
			Type:  &ast.Ident{Name: "int"},
		}
	}
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: name},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: params,
			},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}
}

// createTestAnalyzer создает анализатор с заданными функциями
func createTestAnalyzer(funcs map[string]*ast.FuncDecl) *ParamAnalyzer {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	// Для каждого параметра каждой функции добавим тип int в info.Defs
	for _, f := range funcs {
		if f.Type != nil && f.Type.Params != nil {
			for _, field := range f.Type.Params.List {
				for _, name := range field.Names {
					info.Defs[name] = types.NewVar(0, nil, name.Name, types.Typ[types.Int])
				}
			}
		}
	}
	analyzer := &ParamAnalyzer{
		// lowers: make(map[string][]string),
		all:  funcs,
		info: info,
	}
	return analyzer
}

func TestRecurseCheckDeep(t *testing.T) {
	// Создаем тестовые функции
	funcA := createTestFuncDecl("a", 3)
	funcB := createTestFuncDecl("b", 3)
	funcC := createTestFuncDecl("c", 3)

	// Создаем анализатор
	analyzer := createTestAnalyzer(map[string]*ast.FuncDecl{
		"a": funcA,
		"b": funcB,
		"c": funcC,
	})

	// Создаем тестовый pass
	pass := &analysis.Pass{
		Fset:  token.NewFileSet(),
		Files: []*ast.File{},
		Report: func(d analysis.Diagnostic) {
			t.Logf("Report: %s", d.Message)
		},
	}

	tests := []struct {
		name     string
		call     *ast.CallExpr
		lower    *ast.FuncDecl
		args     map[string]int
		depth    int
		stack    []string
		expected bool
	}{
		{
			name:     "Cyclic call",
			call:     &ast.CallExpr{},
			lower:    funcA,
			args:     make(map[string]int),
			depth:    1,
			stack:    []string{"a"},
			expected: false,
		},
		{
			name:     "Max recursion depth",
			call:     &ast.CallExpr{},
			lower:    funcB,
			args:     make(map[string]int),
			depth:    1,
			stack:    []string{"a"},
			expected: false,
		},
		{
			name:     "Insufficient parameters",
			call:     &ast.CallExpr{},
			lower:    funcB,
			args:     map[string]int{"int": 1},
			depth:    1,
			stack:    []string{"a"},
			expected: false,
		},
		{
			name:     "Successful diagnostic",
			call:     &ast.CallExpr{},
			lower:    funcB,
			args:     map[string]int{"int": 3},
			depth:    0,
			stack:    []string{"a"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.recurseCheckDeep(pass, tt.lower, tt.args, tt.depth, tt.stack)
			hasMessages := len(result.callStack) > 0
			if hasMessages != tt.expected {
				t.Errorf("Expected %v, got %v (messages: %v)", tt.expected, hasMessages, result)
			}
		})
	}
}
