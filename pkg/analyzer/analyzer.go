package analyzer

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const minParams = 3

type (
	myAnalizer struct {
		// Для каждой функции - информация о том какие ф-ии из нее зовутся
		lowers map[string][]string
		// Для каждого имени функции(с именем приемника если есть) - место где она определена
		all map[string]*ast.FuncDecl
	}
)

func (m *myAnalizer) run(pass *analysis.Pass) (any, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	declFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Nodes(declFilter, m.addNodeDecls(pass))
	inspector.Nodes(declFilter, m.getProcessSingleFuncDeclCallback(pass))
	// Запишем список аргументов текущей функции
	// args := initArgsMap(params)
	// fmt.Fprintln(out, "args:", formatArgs(args))

	// пройдемся вглубь по стэку от текущей функции
	// return m.recurseCheckDeep(node, pass, 0, k, args, []*ast.FuncDecl{funcDecl})

	return nil, nil
}

func funcDeclToKey(f *ast.FuncDecl) string {
	k := ""
	if f.Recv != nil {
		k += types.ExprString(f.Recv.List[0].Type)
	}
	return k + f.Name.Name
}

func callExprToKey(f *ast.CallExpr) (string, bool) {
	switch sel := f.Fun.(type) {
	case *ast.SelectorExpr: // foo.ReadFile
		return types.ExprString(sel.X) + " " + sel.Sel.Name, true
	case *ast.Ident: // ReadFile
		return sel.Name, true
	}

	return "unknown", false
}

func (m *myAnalizer) addNodeDecls(*analysis.Pass) func(node ast.Node, _ bool) bool {
	return func(node ast.Node, _ bool) bool {
		funcDecl := node.(*ast.FuncDecl)
		params := funcDecl.Type.Params.List
		if len(params) < minParams {
			return true
		}

		k := funcDeclToKey(funcDecl)
		m.all[k] = funcDecl
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				lowerK, ok := callExprToKey(n)
				if !ok {
					return false
				}
				m.lowers[k] = append(m.lowers[k], lowerK)
			}
			return true
		})
		return true
	}
}

func (m *myAnalizer) getProcessSingleFuncDeclCallback(pass *analysis.Pass) func(node ast.Node, _ bool) bool {
	return func(node ast.Node, _ bool) bool {
		funcDecl := node.(*ast.FuncDecl)
		params := funcDecl.Type.Params.List
		if len(params) < minParams {
			return true
		}

		k := funcDeclToKey(funcDecl)
		for _, lowerK := range m.lowers[k] {
		}
		args := initArgsMap(params)
		stack := []*ast.FuncDecl{funcDecl}
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				lowerK, ok := callExprToKey(n)
				if !ok {
					return false
				}

				lowerFunc, ok := m.all[lowerK]
				if !ok {
					return false
				}

				return m.recurseCheckDeep(node, pass, 2, args, stack, lowerFunc)
			}
			return true
		})
		// TODO определять не по повторяющимся именам, а напрямую по переменным
		return true
	}
}

func (m *myAnalizer) recurseCheckDeep(node ast.Node, pass *analysis.Pass, depth int, args set[string], stack []*ast.FuncDecl, lower *ast.FuncDecl) bool {
	depth--
	intersectedArgs := m.modifyArgsByLower(args, lower)
	// Если количество повторяющихся аргументов уменьшилось до n или менее - закончим проверку
	if len(intersectedArgs) < minParams {
		return true
	}

	stack = append(stack, lower)
	if depth == 0 {
		pass.Reportf(node.Pos(), "make struct with arguments: %s, for call stack: %s", formatArgs(intersectedArgs), formatStack(stack))
		return false
	}

	ast.Inspect(lower, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			lowerK, ok := callExprToKey(n)
			if !ok {
				return false
			}
			if lower, ok := m.all[lowerK]; ok {
				return m.recurseCheckDeep(node, pass, depth, intersectedArgs, stack, lower)
			}
		}
		return true
	})

	return false
}

func initArgsMap(params []*ast.Field) set[string] {
	args := make(set[string], len(params))
	for _, v := range params {
		// for _, name := range v.Names {
		// 'a int'
		// 'b myStructTypeABC'
		// fullName := name.Name + " " + types.ExprString(v.Type)
		fullName := types.ExprString(v.Type)
		args[fullName] = struct{}{}
		// }
	}
	return args
}

func (m *myAnalizer) modifyArgsByLower(args set[string], lower *ast.FuncDecl) set[string] {
	lowerArgs := initArgsMap(lower.Type.Params.List)
	return args.Intersection(lowerArgs)
}

func formatArgs(args map[string]struct{}) string {
	b := strings.Builder{}
	b.WriteRune('[')
	for k := range args {
		b.WriteString(k)
		b.WriteRune(',')
		b.WriteRune(' ')
	}
	b.WriteRune(']')
	return b.String()
}

func formatStack(stack []*ast.FuncDecl) string {
	b := strings.Builder{}
	for _, node := range stack {
		b.WriteString(node.Name.Name)
		b.WriteRune('>')
	}
	return b.String()
}

func Analyzer() *analysis.Analyzer {
	m := myAnalizer{
		lowers: map[string][]string{},
		all:    map[string]*ast.FuncDecl{},
	}
	return &analysis.Analyzer{
		Name:     "analyzer",
		Doc:      "analyzer",
		Run:      m.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}
