package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ParamAnalyzer описывает анализатор, который проверяет цепочки вызовов функций
// и предлагает создать структуру для групп параметров, которые передаются через цепочку
type ParamAnalyzer struct {
	// all хранит все объявления функций
	all map[string]*ast.FuncDecl
	ma  sync.RWMutex
	// info хранит информацию о типах
	info *types.Info
	// results хранит все подходящие цепочки
	results []chainResult
	// minRequiredParams определяет минимальное количество параметров,
	// при котором анализатор начинает проверку на необходимость создания структуры
	minRequiredParams int
	// maxRecursionDepth определяет максимальную глубину рекурсии при анализе цепочки вызовов
	maxRecursionDepth int
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

	inspector.Nodes(declFilter, m.addNodeDecls())
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

func (m *ParamAnalyzer) addNodeDecls() func(node ast.Node, push bool) bool {
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
		if totalParams < m.minRequiredParams {
			return true
		}

		k := m.funcDeclToKey(funcDecl)
		m.ma.Lock()
		m.all[k] = funcDecl
		m.ma.Unlock()
		return true
	}
}

type chainResult struct {
	callStack []string
	msg       string
	leafFunc  *ast.FuncDecl // конечная функция цепочки
}

func (m *ParamAnalyzer) recurseCheckDeep(pass *analysis.Pass, currentFunc *ast.FuncDecl, args map[string]int, depth int, callStack []string) chainResult {
	if depth > m.maxRecursionDepth {
		return chainResult{}
	}

	currentFuncKey := m.funcDeclToKey(currentFunc)
	newCallStack := append(callStack, currentFuncKey)

	// Модифицируем аргументы на основе нижней функции
	currentFuncArgs := make(map[string]int)
	for _, param := range currentFunc.Type.Params.List {
		if t := m.info.TypeOf(param.Type); t != nil {
			currentFuncArgs[t.String()] += max(len(param.Names), 1) // обработка безымянных параметров
			continue
		}

		for _, name := range param.Names {
			if name == nil {
				continue
			}

			param := m.info.ObjectOf(name)
			if param == nil {
				continue
			}

			if paramType := param.Type(); paramType != nil {
				currentFuncArgs[paramType.String()]++
			}
		}
	}

	// Пересекаем аргументы
	intersectedArgs := make(map[string]int)
	for t, count := range args {
		if lowerCount, ok := currentFuncArgs[t]; ok {
			intersectedArgs[t] = min(count, lowerCount)
		}
	}

	if totalParams(intersectedArgs) < 3 {
		return chainResult{}
	}

	// Собираем все вызовы в текущей функции
	var calls []*ast.CallExpr
	ast.Inspect(currentFunc, func(node ast.Node) bool {
		if callExpr, ok := node.(*ast.CallExpr); ok {
			calls = append(calls, callExpr)
		}
		return true
	})

	// Если нет вызовов и это конечная функция в цепочке
	if len(calls) == 0 {
		argsStr := make([]string, 0, len(intersectedArgs))
		for t, n := range intersectedArgs {
			for range n {
				argsStr = append(argsStr, t)
			}
		}
		sort.Strings(argsStr)
		callStackStr := strings.Join(newCallStack, " -> ")
		msg := fmt.Sprintf("make struct with arguments: %s, for call stack: %s", strings.Join(argsStr, ", "), callStackStr)
		return chainResult{
			callStack: newCallStack,
			msg:       msg,
			leafFunc:  currentFunc,
		}
	}

	// Для каждого вызова создаем отдельную цепочку
	chains := m.getChains(pass, calls, newCallStack, intersectedArgs, depth)
	if len(chains) <= 0 {
		return chainResult{}
	}

	// Если есть результаты, возвращаем только самую длинную цепочку
	maxResult := chains[0]
	for _, res := range chains[1:] {
		if len(res.callStack) > len(maxResult.callStack) {
			maxResult = res
		}
	}

	return maxResult
}

func (m *ParamAnalyzer) getChains(pass *analysis.Pass, calls []*ast.CallExpr, newCallStack []string, intersectedArgs map[string]int, depth int) []chainResult {
	chains := make([]chainResult, 0, len(calls))
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

		m.ma.RLock()
		calledFunc, ok := m.all[calledKey]
		m.ma.RUnlock()
		if !ok || calledFunc == nil {
			continue
		}

		res := m.recurseCheckDeep(pass, calledFunc, intersectedArgs, depth+1, newCallStack)
		if res.msg == "" {
			continue
		}

		chains = append(chains, res)
	}
	return chains
}

// initArgsMap возвращает map[string]int, где ключ — тип, а значение — количество параметров этого типа
func (m *ParamAnalyzer) initArgsMap(params []*ast.Field) map[string]int {
	args := make(map[string]int)
	for _, field := range params {
		var typeStr string
		if t := m.info.TypeOf(field.Type); t != nil {
			typeStr = t.String()
		} else {
			typeStr = types.ExprString(field.Type) // fallback
		}

		count := 1
		if len(field.Names) > 0 {
			count = len(field.Names)
		}
		args[typeStr] += count
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

// Analyzer создает новый анализатор параметров с значениями по умолчанию
func Analyzer() *analysis.Analyzer {
	return AnalyzerWithConfig(2, 10) // default values
}

// AnalyzerWithConfig создает новый анализатор параметров с указанными конфигурационными значениями
func AnalyzerWithConfig(minRequiredParams, maxRecursionDepth int) *analysis.Analyzer {
	m := &ParamAnalyzer{
		all:               make(map[string]*ast.FuncDecl),
		minRequiredParams: minRequiredParams,
		maxRecursionDepth: maxRecursionDepth,
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
		if totalParams < m.minRequiredParams {
			return true
		}

		k := m.funcDeclToKey(funcDecl)
		args := m.initArgsMap(params)

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
			m.ma.RLock()
			lowerFunc, ok := m.all[lowerK]
			m.ma.RUnlock()
			if !ok {
				continue
			}
			// Передаем стек, начинающийся с текущей функции (корня)
			res := m.recurseCheckDeep(pass, lowerFunc, args, 1, []string{k})
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
		if slices.ContainsFunc(result, func(v chainResult) bool { return isSubChainOf(chain.callStack, v.callStack) }) {
			continue
		}

		// Уникальность по callStack
		key := strings.Join(chain.callStack, "->")
		if _, exists := seen[key]; !exists {
			result = append(result, chain)
			seen[key] = struct{}{}
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
		for j := range sub {
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
