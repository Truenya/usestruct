package testcase

// Цепочка 1: a -> b -> c -> d -> e -> f (самая длинная)
func a(x, y, z int) { b(x, y, z) }
func b(x, y, z int) { c(x, y, z) }
func c(x, y, z int) { d(x, y, z) }
func d(x, y, z int) { e(x, y, z) }
func e(x, y, z int) { f(x, y, z) }
func f(x, y, z int) {} // want "make struct with arguments: int, int, int, for call stack: a -> b -> c -> d -> e -> f"

// Цепочка 2: g -> h -> i (короткая)
func g(x, y, z int) { h(x, y, z) }
func h(x, y, z int) { i(x, y, z) }
func i(x, y, z int) {} // want "make struct with arguments: int, int, int, for call stack: g -> h -> i"

// Цепочка 3: j -> k -> l -> m (средняя)
func j(x, y, z int) { k(x, y, z) }
func k(x, y, z int) { l(x, y, z) }
func l(x, y, z int) { m(x, y, z) }
func m(x, y, z int) {} // want "make struct with arguments: int, int, int, for call stack: j -> k -> l -> m"
