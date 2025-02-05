package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// Mocker предоставляет методы для создания тестовых данных
type Mocker struct {
	info *types.Info
}

// NewMocker создает новый Mocker
func NewMocker() *Mocker {
	return &Mocker{
		info: &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Implicits:  make(map[ast.Node]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Scopes:     make(map[ast.Node]*types.Scope),
			InitOrder:  make([]*types.Initializer, 0),
		},
	}
}

// AddType добавляет информацию о типе в мокер
func (m *Mocker) AddType(expr ast.Expr, typ types.Type) {
	m.info.Types[expr] = types.TypeAndValue{
		Type: typ,
	}
}

// NewNamedType создает новый именованный тип
func (m *Mocker) NewNamedType(name string) *types.Named {
	return types.NewNamed(
		types.NewTypeName(token.NoPos, nil, name, nil),
		types.NewStruct(nil, nil),
		nil,
	)
}

// SetupMethodTypes настраивает типы для метода и его вызова
func (m *Mocker) SetupMethodTypes(funcDecl *ast.FuncDecl, callExpr *ast.CallExpr, isPointer bool) {
	if funcDecl == nil || callExpr == nil {
		return
	}

	// Создаем тип для receiver
	typeName := "Test" // Можно сделать параметром, если нужно
	testType := m.NewNamedType(typeName)

	// Настраиваем тип для receiver в объявлении функции
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if isPointer {
			if star, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
				m.AddType(funcDecl.Recv.List[0].Type, types.NewPointer(testType))
				m.AddType(star.X, testType)
			}
		} else {
			m.AddType(funcDecl.Recv.List[0].Type, testType)
		}
	}

	// Настраиваем тип для переменной в вызове
	if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if isPointer {
				m.AddType(ident, types.NewPointer(testType))
			} else {
				m.AddType(ident, testType)
			}
		}
	}
}

// ParseAndSetupTypes парсит код и настраивает типы для методов
func (m *Mocker) ParseAndSetupTypes(code string) (*ast.File, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		return nil, err
	}

	// Находим все методы и их вызовы
	var funcDecls []*ast.FuncDecl
	var callExprs []*ast.CallExpr

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Recv != nil {
				funcDecls = append(funcDecls, node)
			}
		case *ast.CallExpr:
			if _, ok := node.Fun.(*ast.SelectorExpr); ok {
				callExprs = append(callExprs, node)
			}
		}
		return true
	})

	// Настраиваем типы для каждого метода и его вызова
	for i, funcDecl := range funcDecls {
		if i < len(callExprs) {
			// Определяем, является ли receiver указателем
			isPointer := false
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if _, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
					isPointer = true
				}
			}
			m.SetupMethodTypes(funcDecl, callExprs[i], isPointer)
		}
	}

	return file, nil
}

// Analyzer возвращает ParamAnalyzer с настроенными тестовыми данными
func (m *Mocker) Analyzer() *ParamAnalyzer {
	return &ParamAnalyzer{
		all:  make(map[string]*ast.FuncDecl),
		info: m.info,
	}
}
