

package analyzer

import (
	"reflect"
	"testing"
)

func TestFilterMaxChains(t *testing.T) {
	t.Run("empty chains returns nil", func(t *testing.T) {
		result := filterMaxChains([]chainResult{})
		if result != nil {
			t.Errorf("expected nil for empty input, got %v", result)
		}
	})

	t.Run("filters out empty messages", func(t *testing.T) {
		chains := []chainResult{
			{msg: ""},
			{msg: "test message"},
		}
		result := filterMaxChains(chains)
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].msg != "test message" {
			t.Errorf("expected 'test message', got '%s'", result[0].msg)
		}
	})

	t.Run("keeps only longest unique chains", func(t *testing.T) {
		// Create test chains
		longChain := chainResult{
			callStack: []string{"a", "b", "c", "d"},
			msg:       "longest chain",
		}
		mediumChain := chainResult{
			callStack: []string{"a", "b", "c"},
			msg:       "medium chain",
		}
		shortChain := chainResult{
			callStack: []string{"a", "b"},
			msg:       "short chain",
		}

		chains := []chainResult{shortChain, mediumChain, longChain}

		result := filterMaxChains(chains)
		if len(result) != 1 {
			t.Errorf("expected 1 result (longest chain), got %d", len(result))
		}
		if !reflect.DeepEqual(result[0], longChain) {
			t.Errorf("expected longest chain, got %v", result[0])
		}
	})

	t.Run("handles duplicate call stacks", func(t *testing.T) {
		chain1 := chainResult{
			callStack: []string{"a", "b", "c"},
			msg:       "first message",
		}
		chain2 := chainResult{
			callStack: []string{"a", "b", "c"},
			msg:       "second message",
		}

		chains := []chainResult{chain1, chain2}
		result := filterMaxChains(chains)

		if len(result) != 1 {
			t.Errorf("expected only one unique result, got %d", len(result))
		}
	})

	t.Run("handles sub-chains correctly", func(t *testing.T) {
		parentChain := chainResult{
			callStack: []string{"a", "b", "c", "d"},
			msg:       "parent message",
		}
		subChain := chainResult{
			callStack: []string{"a", "b", "c"},
			msg:       "sub-chain message",
		}

		chains := []chainResult{subChain, parentChain}
		result := filterMaxChains(chains)

		if len(result) != 1 {
			t.Errorf("expected only one result (parent chain), got %d", len(result))
		}
		if !reflect.DeepEqual(result[0], parentChain) {
			t.Errorf("expected parent chain, got %v", result[0])
		}
	})
}

func TestIsSubChainOf(t *testing.T) {
	t.Run("empty sub-chain returns true", func(t *testing.T) {
		main := []string{"a", "b", "c"}
		sub := []string{}

		result := isSubChainOf(sub, main)
		if result != false {
			t.Errorf("expected true for empty sub-chain")
		}
	})

	t.Run("sub-chain longer than main returns false", func(t *testing.T) {
		sub := []string{"a", "b", "c", "d"}
		main := []string{"a", "b", "c"}

		result := isSubChainOf(sub, main)
		if result {
			t.Errorf("expected false for sub-chain longer than main")
		}
	})

	t.Run("identical chains returns false", func(t *testing.T) {
		chain := []string{"a", "b", "c"}

		result := isSubChainOf(chain, chain)
		if result {
			t.Errorf("expected false for identical chains")
		}
	})

	t.Run("valid sub-chain returns true", func(t *testing.T) {
		main := []string{"a", "b", "c", "d", "e"}
		sub := []string{"b", "c", "d"}

		result := isSubChainOf(sub, main)
		if !result {
			t.Errorf("expected true for valid sub-chain")
		}
	})

	t.Run("non-continuous sub-chain returns false", func(t *testing.T) {
		main := []string{"a", "b", "c", "d", "e"}
		sub := []string{"a", "c", "e"}

		result := isSubChainOf(sub, main)
		if result {
			t.Errorf("expected false for non-continuous sub-chain")
		}
	})

	t.Run("sub-chain at start returns true", func(t *testing.T) {
		main := []string{"a", "b", "c", "d"}
		sub := []string{"a", "b"}

		result := isSubChainOf(sub, main)
		if !result {
			t.Errorf("expected true for sub-chain at start")
		}
	})

	t.Run("sub-chain at end returns true", func(t *testing.T) {
		main := []string{"a", "b", "c", "d"}
		sub := []string{"c", "d"}

		result := isSubChainOf(sub, main)
		if !result {
			t.Errorf("expected true for sub-chain at end")
		}
	})
}

