package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// MinRequiredParams определяет минимальное количество параметров,
// при котором анализатор начинает проверку на необходимость создания структуры
const MinRequiredParams = 3

// MaxRecursionDepth определяет максимальную глубину рекурсии при анализе цепочки вызовов
const MaxRecursionDepth = 10

// ParamAnalyzer описывает анализатор, который проверяет цепочки вызовов функций
// и предлагает создать структуру для групп параметров, которые передаются через цепочку
type ParamAnalyzer struct {
	// lowers хранит карту вызовов функций для каждой функции
	lowers map[string][]string
	// all хранит все объявления функций
	all map[string]*ast.FuncDecl
	// info хранит информацию о типах
	info *types.Info
	// results хранит все подходящие цепочки
	results []chainResult
}

// run выполняет анализ кода
func (m *ParamAnalyzer) run(pass *analysis.Pass) (any, error) {
	m.info = pass.TypesInfo
	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("failed to get inspector from pass")
	}

	declFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Nodes(declFilter, m.addNodeDecls(pass))
	inspector.Nodes(declFilter, m.getProcessSingleFuncDeclCallback(pass))

	// Фильтруем только максимальные цепочки (не вложенные)
	maxChains := filterMaxChains(m.results)
	for _, res := range maxChains {
		if res.msg != "" && res.leafFunc != nil {
			pass.Reportf(res.leafFunc.Pos(), res.msg)
		}
	}

	return nil, nil
}

// funcDeclToKey преобразует объявление функции в ключ
func (m *ParamAnalyzer) funcDeclToKey(f *ast.FuncDecl) string {
	if f == nil {
		return ""
	}

	k := ""
	if f.Recv != nil && len(f.Recv.List) > 0 && f.Recv.List[0].Type != nil {
		// Получаем тип из types.Info
		if t := m.info.TypeOf(f.Recv.List[0].Type); t != nil {
			// Убираем указатель, если есть
			if ptr, ok := t.(*types.Pointer); ok {
				t = ptr.Elem()
			}
			// Используем только имя типа
			if named, ok := t.(*types.Named); ok {
				k = named.Obj().Name() + "."
			}
		}
	}
	return k + f.Name.Name
}

// callExprToKey преобразует выражение вызова в ключ
func (m *ParamAnalyzer) callExprToKey(f *ast.CallExpr) (string, bool) {
	if f == nil || f.Fun == nil {
		return "", false
	}

	switch sel := f.Fun.(type) {
	case *ast.SelectorExpr:
		if sel.X == nil || sel.Sel == nil {
			return "", false
		}
		// Получаем тип из types.Info
		if t := m.info.TypeOf(sel.X); t != nil {
			// Убираем указатель, если есть
			if ptr, ok := t.(*types.Pointer); ok {
				t = ptr.Elem()
			}
			// Используем только имя типа
			if named, ok := t.(*types.Named); ok {
				return named.Obj().Name() + "." + sel.Sel.Name, true
			}
		}
		return "", false
	case *ast.Ident:
		return sel.Name, true
	}

	return "", false
}

func (m *ParamAnalyzer) addNodeDecls(pass *analysis.Pass) func(node ast.Node, push bool) bool {
	return func(node ast.Node, push bool) bool {
		if !push {
			return true
		}

		funcDecl := node.(*ast.FuncDecl)
		params := funcDecl.Type.Params.List
		totalParams := 0
		for _, param := range params {
			if len(param.Names) > 0 {
				totalParams += len(param.Names)
			} else {
				totalParams++
			}
		}
		if totalParams < MinRequiredParams {
			return true
		}

		k := m.funcDeclToKey(funcDecl)
		m.all[k] = funcDecl
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				lowerK, ok := m.callExprToKey(n)
				if !ok {
					return true
				}
				m.lowers[k] = append(m.lowers[k], lowerK)
			}
			return true
		})
		return true
	}
}

type chainResult struct {
	callStack []string
	msg       string
	leafFunc  *ast.FuncDecl // конечная функция цепочки
}

func (m *ParamAnalyzer) recurseCheckDeep(pass *analysis.Pass, call *ast.CallExpr, lowerFunc *ast.FuncDecl, args map[string]int, depth int, callStack []string) chainResult {
	if depth > MaxRecursionDepth {
		return chainResult{}
	}

	lowerKey := m.funcDeclToKey(lowerFunc)
	newCallStack := append(callStack, lowerKey)

	// Модифицируем аргументы на основе нижней функции
	lowerFuncArgs := make(map[string]int)
	for _, param := range lowerFunc.Type.Params.List {
		for _, name := range param.Names {
			paramType := m.info.ObjectOf(name).Type().String()
			lowerFuncArgs[paramType]++
		}
	}

	// Пересекаем аргументы
	intersectedArgs := make(map[string]int)
	for t, count := range args {
		if lowerCount, ok := lowerFuncArgs[t]; ok {
			intersectedArgs[t] = min(count, lowerCount)
		}
	}

	if totalParams(intersectedArgs) < 3 {
		return chainResult{}
	}

	// Собираем все вызовы в текущей функции
	var calls []*ast.CallExpr
	ast.Inspect(lowerFunc, func(node ast.Node) bool {
		if callExpr, ok := node.(*ast.CallExpr); ok {
			calls = append(calls, callExpr)
		}
		return true
	})

	// Если нет вызовов и это конечная функция в цепочке
	if len(calls) == 0 {
		argsStr := make([]string, 0, len(intersectedArgs))
		for t, n := range intersectedArgs {
			for i := 0; i < n; i++ {
				argsStr = append(argsStr, t)
			}
		}
		sort.Strings(argsStr)
		callStackStr := strings.Join(newCallStack, " -> ")
		msg := fmt.Sprintf("make struct with arguments: %s, for call stack: %s", strings.Join(argsStr, ", "), callStackStr)
		return chainResult{
			callStack: newCallStack,
			msg:       msg,
			leafFunc:  lowerFunc,
		}
	}

	// Для каждого вызова создаем отдельную цепочку
	var results []chainResult
	for _, callExpr := range calls {
		calledKey, ok := m.callExprToKey(callExpr)
		if !ok {
			continue
		}

		// Проверяем, сколько раз функция уже встречалась в стеке
		foundCount := 0
		for _, f := range newCallStack {
			if f == calledKey {
				foundCount++
			}
		}
		// Разрешаем не более 2 вхождений одной функции в стек
		if foundCount >= 2 {
			continue
		}

		calledFunc, ok := m.all[calledKey]
		if ok && calledFunc != nil {
			res := m.recurseCheckDeep(pass, callExpr, calledFunc, intersectedArgs, depth+1, newCallStack)
			if res.msg != "" {
				results = append(results, res)
			}
		}
	}

	// Если есть результаты, возвращаем только самую длинную цепочку
	if len(results) > 0 {
		maxResult := results[0]
		for _, res := range results[1:] {
			if len(res.callStack) > len(maxResult.callStack) {
				maxResult = res
			}
		}
		return maxResult
	}

	return chainResult{}
}

// initArgsMap теперь возвращает map[string]int, где ключ — тип, а значение — количество параметров этого типа
func initArgsMap(params []*ast.Field) map[string]int {
	args := make(map[string]int)
	for _, v := range params {
		fullName := types.ExprString(v.Type)
		count := 1
		if len(v.Names) > 0 {
			count = len(v.Names)
		}
		args[fullName] += count
	}
	return args
}

// Считает общее количество параметров (по всем типам)
func totalParams(args map[string]int) int {
	sum := 0
	for _, v := range args {
		sum += v
	}
	return sum
}

// Analyzer создает новый анализатор параметров
func Analyzer() *analysis.Analyzer {
	m := &ParamAnalyzer{
		lowers: make(map[string][]string),
		all:    make(map[string]*ast.FuncDecl),
	}
	return &analysis.Analyzer{
		Name:     "paramStructAnalyzer",
		Doc:      "suggests to make struct for group of arguments passed through function chain",
		Run:      m.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func (m *ParamAnalyzer) getProcessSingleFuncDeclCallback(pass *analysis.Pass) func(node ast.Node, push bool) bool {
	return func(node ast.Node, push bool) bool {
		if !push {
			return true
		}

		funcDecl := node.(*ast.FuncDecl)
		params := funcDecl.Type.Params.List
		totalParams := 0
		for _, param := range params {
			if len(param.Names) > 0 {
				totalParams += len(param.Names)
			} else {
				totalParams++
			}
		}
		if totalParams < MinRequiredParams {
			return true
		}

		k := m.funcDeclToKey(funcDecl)
		args := initArgsMap(params)

		// Собираем все вызовы в текущей функции
		var calls []*ast.CallExpr
		ast.Inspect(funcDecl, func(node ast.Node) bool {
			if callExpr, ok := node.(*ast.CallExpr); ok {
				calls = append(calls, callExpr)
			}
			return true
		})

		// Для каждого вызова создаем отдельную цепочку
		for _, callExpr := range calls {
			lowerK, ok := m.callExprToKey(callExpr)
			if !ok {
				continue
			}
			lowerFunc, ok := m.all[lowerK]
			if !ok {
				continue
			}
			// Передаем стек, начинающийся с текущей функции (корня)
			res := m.recurseCheckDeep(pass, callExpr, lowerFunc, args, 1, []string{k})
			if res.msg != "" && res.leafFunc != nil {
				m.results = append(m.results, res)
			}
		}

		return true
	}
}

// filterMaxChains оставляет только уникальные максимальные цепочки (без вложенных и дублей по callStack)
func filterMaxChains(chains []chainResult) []chainResult {
	if len(chains) == 0 {
		return nil
	}

	// Сортируем цепочки по длине (от длинных к коротким)
	sort.Slice(chains, func(i, j int) bool {
		return len(chains[i].callStack) > len(chains[j].callStack)
	})

	var result []chainResult
	seen := make(map[string]struct{})

	for _, chain := range chains {
		if chain.msg == "" {
			continue
		}

		// Проверяем, не является ли эта цепочка подцепочкой уже обработанной
		isSubChain := false
		for _, prevChain := range result {
			if isSubChainOf(chain.callStack, prevChain.callStack) {
				isSubChain = true
				break
			}
		}

		if !isSubChain {
			// Уникальность по callStack
			key := strings.Join(chain.callStack, "->")
			if _, exists := seen[key]; !exists {
				result = append(result, chain)
				seen[key] = struct{}{}
			}
		}
	}

	return result
}

func isSubChainOf(sub, main []string) bool {
	if len(sub) >= len(main) {
		return false
	}

	// Ищем подцепочку в основной цепочке
	for i := 0; i <= len(main)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if sub[j] != main[i+j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
