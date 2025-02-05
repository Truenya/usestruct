package analyzer

type set[T comparable] map[T]struct{}

// NewSet создаёт новое множество.
func NewSet[T comparable](ss ...T) set[T] {
	set := make(set[T], len(ss))

	for _, s := range ss {
		set[s] = struct{}{}
	}

	return set
}

// Intersection возвращает пересечение текущего и переданного множества.
func (s set[T]) Intersection(other set[T]) set[T] {
	if len(s) == 0 || len(other) == 0 {
		return nil
	}

	ret := NewSet[T]()

	for v := range other {
		if s.Has(v) {
			ret.Add(v)
		}
	}

	return ret
}

// Has проверят наличие элемента в множестве.
func (s set[T]) Has(item T) bool {
	if s == nil {
		return false
	}

	_, ok := s[item]
	return ok
}

// Add добавляет новое значение в множество.
func (s set[T]) Add(item T) {
	if s == nil {
		return
	}

	s[item] = struct{}{}
}
