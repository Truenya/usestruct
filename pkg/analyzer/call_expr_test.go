package analyzer

import (
	"go/ast"
	"testing"
)

func TestCallExprToKey(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Mocker) *ast.CallExpr
		expected string
		ok       bool
	}{
		{
			name: "simple function call",
			setup: func(m *Mocker) *ast.CallExpr {
				file, err := m.ParseAndSetupTypes(`
					package p
					func Test() {}
					func main() { Test() }
				`)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}
				var callExpr *ast.CallExpr
				ast.Inspect(file, func(n ast.Node) bool {
					if ce, ok := n.(*ast.CallExpr); ok {
						callExpr = ce
						return false
					}
					return true
				})
				return callExpr
			},
			expected: "Test",
			ok:       true,
		},
		{
			name: "method call with pointer receiver",
			setup: func(m *Mocker) *ast.CallExpr {
				file, err := m.ParseAndSetupTypes(`
					package p
					type Test struct{}
					func (t *Test) Test() {}
					func main() {
						var t *Test
						t.Test()
					}
				`)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}
				var callExpr *ast.CallExpr
				ast.Inspect(file, func(n ast.Node) bool {
					if ce, ok := n.(*ast.CallExpr); ok {
						callExpr = ce
						return false
					}
					return true
				})
				return callExpr
			},
			expected: "Test.Test",
			ok:       true,
		},
		{
			name: "method call with value receiver",
			setup: func(m *Mocker) *ast.CallExpr {
				file, err := m.ParseAndSetupTypes(`
					package p
					type Test struct{}
					func (t Test) Test() {}
					func main() {
						var t Test
						t.Test()
					}
				`)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}
				var callExpr *ast.CallExpr
				ast.Inspect(file, func(n ast.Node) bool {
					if ce, ok := n.(*ast.CallExpr); ok {
						callExpr = ce
						return false
					}
					return true
				})
				return callExpr
			},
			expected: "Test.Test",
			ok:       true,
		},
		{
			name: "nil call expression",
			setup: func(m *Mocker) *ast.CallExpr {
				return nil
			},
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocker := NewMocker()
			callExpr := tt.setup(mocker)
			analyzer := mocker.Analyzer()
			got, ok := analyzer.callExprToKey(callExpr)
			if ok != tt.ok {
				t.Errorf("callExprToKey() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("callExprToKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFuncDeclAndCallExprKeysMatch(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Mocker) (*ast.FuncDecl, *ast.CallExpr)
	}{
		{
			name: "method with pointer receiver",
			setup: func(m *Mocker) (*ast.FuncDecl, *ast.CallExpr) {
				file, err := m.ParseAndSetupTypes(`
					package p
					type Test struct{}
					func (t *Test) Test() {}
					func main() {
						var t *Test
						t.Test()
					}
				`)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}

				var funcDecl *ast.FuncDecl
				var callExpr *ast.CallExpr

				ast.Inspect(file, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncDecl:
						if node.Recv != nil {
							funcDecl = node
						}
					case *ast.CallExpr:
						if _, ok := node.Fun.(*ast.SelectorExpr); ok {
							callExpr = node
						}
					}
					return true
				})

				return funcDecl, callExpr
			},
		},
		{
			name: "method with value receiver",
			setup: func(m *Mocker) (*ast.FuncDecl, *ast.CallExpr) {
				file, err := m.ParseAndSetupTypes(`
					package p
					type Test struct{}
					func (t Test) Test() {}
					func main() {
						var t Test
						t.Test()
					}
				`)
				if err != nil {
					t.Fatalf("failed to parse code: %v", err)
				}

				var funcDecl *ast.FuncDecl
				var callExpr *ast.CallExpr

				ast.Inspect(file, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncDecl:
						if node.Recv != nil {
							funcDecl = node
						}
					case *ast.CallExpr:
						if _, ok := node.Fun.(*ast.SelectorExpr); ok {
							callExpr = node
						}
					}
					return true
				})

				return funcDecl, callExpr
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocker := NewMocker()
			funcDecl, callExpr := tt.setup(mocker)
			if funcDecl == nil || callExpr == nil {
				t.Fatal("failed to find function declaration or call expression")
			}

			analyzer := mocker.Analyzer()
			funcKey := analyzer.funcDeclToKey(funcDecl)
			callKey, ok := analyzer.callExprToKey(callExpr)
			if !ok {
				t.Fatal("failed to get call key")
			}

			if funcKey != callKey {
				t.Errorf("funcDeclToKey() = %v, callExprToKey() = %v, want them to be equal", funcKey, callKey)
			}
		})
	}
}
