package analyzer

// func TestFormatArgs(t *testing.T) {
// 	args := map[string]int{
// 		"int":     1,
// 		"string":  1,
// 		"float64": 1,
// 	}
// 	got := formatArgs(args)

// 	// Проверяем, что все элементы есть в строке, а их количество совпадает
// 	parts := strings.Split(got, ", ")
// 	expectedSet := map[string]int{
// 		"int":     1,
// 		"string":  1,
// 		"float64": 1,
// 	}
// 	if len(parts) != 3 { // 3 элемента: int, string, float64
// 		t.Errorf("formatArgs() = %v, want 3 elements", got)
// 	}
// 	for _, p := range parts {
// 		if _, ok := expectedSet[p]; !ok {
// 			t.Errorf("formatArgs() = %v, unexpected part: %v", got, p)
// 		}
// 	}
// }

// func TestFormatStack(t *testing.T) {
// 	fset := token.NewFileSet()
// 	code := `package test
// func foo() {}
// func bar() {}
// func baz() {}`
// 	file, err := parser.ParseFile(fset, "", code, 0)
// 	if err != nil {
// 		t.Fatalf("failed to parse code: %v", err)
// 	}
// 	stack := make([]*ast.FuncDecl, 0)
// 	for _, decl := range file.Decls {
// 		if fd, ok := decl.(*ast.FuncDecl); ok {
// 			stack = append(stack, fd)
// 		}
// 	}
// 	got := formatStack(stack)
// 	expected := "foo -> bar -> baz"
// 	if got != expected {
// 		t.Errorf("formatStack() = %v, want %v", got, expected)
// 	}
// }
