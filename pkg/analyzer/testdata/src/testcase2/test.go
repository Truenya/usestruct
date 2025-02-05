package testcase2

import (
	"fmt"
)

// Тест 1: Цепочка с разными типами параметров
func processInt(x, y, z int) { processFloat(float64(x), float64(y), float64(z)) }
func processFloat(a, b, c float64) {
	processString(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b), fmt.Sprintf("%v", c))
}
func processString(s1, s2, s3 string) {} // (диагностики не должно быть)

// Тест 2: Цепочка с методами структуры
type Processor struct{}

func (p *Processor) Start(x, y, z int)   { p.Process(x, y, z) }
func (p *Processor) Process(a, b, c int) { p.Finish(a, b, c) }
func (p *Processor) Finish(i, j, k int)  {} // want "make struct with arguments: int, int, int, for call stack: Processor.Start -> Processor.Process -> Processor.Finish"

// Тест 3: Вложенные цепочки (должна быть выбрана только самая длинная)
func alpha(x, y, z int)   { beta(x, y, z) }
func beta(a, b, c int)    { gamma(a, b, c) }
func gamma(i, j, k int)   { delta(i, j, k) }
func delta(p, q, r int)   { epsilon(p, q, r) }
func epsilon(u, v, w int) {} // want "make struct with arguments: int, int, int, for call stack: alpha -> beta -> gamma -> delta -> epsilon"

// Тест 4: Цепочка с условными вызовами
func check(x, y, z int) {
	if x > 0 {
		validate(x, y, z)
	} else {
		verify(x, y, z)
	}
}
func validate(a, b, c int) {
	a--
	check(a, b, c)
}
func verify(i, j, k int)  { confirm(i, j, k) }
func confirm(p, q, r int) {} // want "make struct with arguments: int, int, int, for call stack: validate -> check -> validate -> check -> verify -> confirm"

// Тест 5: Цепочка с разным количеством параметров
func start(x, y, z int) { middle(x, y) }
func middle(a, b int)   { end(a) }
func end(i int)         {} // Не должно быть диагностики, так как меньше 3 параметров

// Тест 6: Цепочка с методами разных структур
type Handler struct{}
type DataProcessor struct{}

func (h *Handler) Init(x, y, z int)           { h.Process(x, y, z) }
func (h *Handler) Process(a, b, c int)        { p := &DataProcessor{}; p.Handle(a, b, c) }
func (p *DataProcessor) Handle(i, j, k int)   { p.Finalize(i, j, k) }
func (p *DataProcessor) Finalize(u, v, w int) {} // want "make struct with arguments: int, int, int, for call stack: Handler.Init -> Handler.Process -> DataProcessor.Handle -> DataProcessor.Finalize"
