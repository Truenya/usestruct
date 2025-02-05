
package analyzer

import (
	"testing"
)

func TestSet(t *testing.T) {
	t.Run("NewSet creates a non-nil set", func(t *testing.T) {
		s := NewSet(1, 2, 3)
		if s == nil {
			t.Fatalf("expected non-nil set")
		}
	})

	t.Run("Has returns true for existing items", func(t *testing.T) {
		s := NewSet(1, 2, 3)
		if !s.Has(2) {
			t.Errorf("expected Has to return true for item 2")
		}
	})

	t.Run("Has returns false for non-existent items", func(t *testing.T) {
		s := NewSet(1, 2, 3)
		if s.Has(4) {
			t.Errorf("expected Has to return false for item 4")
		}
	})

	t.Run("Add adds new items to the set", func(t *testing.T) {
		s := NewSet[int]()
		s.Add(1)
		if !s.Has(1) {
			t.Errorf("expected Has to return true after adding item 1")
		}
	})

	t.Run("Intersection returns common elements", func(t *testing.T) {
		s1 := NewSet(1, 2, 3)
		s2 := NewSet(2, 3, 4)

		result := s1.Intersection(s2)
		if len(result) != 2 {
			t.Errorf("expected intersection to have 2 elements, got %d", len(result))
		}

		if !result.Has(2) || !result.Has(3) {
			t.Errorf("expected intersection to contain elements 2 and 3")
		}
	})

	t.Run("Intersection with empty set returns nil", func(t *testing.T) {
		s1 := NewSet(1, 2, 3)
		var s2 set[int]

		result := s1.Intersection(s2)
		if result != nil {
			t.Errorf("expected intersection to be nil for empty sets")
		}
	})

	t.Run("Has returns false for nil set", func(t *testing.T) {
		var s set[int]
		if s.Has(1) {
			t.Errorf("expected Has to return false for nil set")
		}
	})
}
