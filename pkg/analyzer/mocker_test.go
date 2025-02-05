package analyzer

import (
	"testing"
)

func TestMocker(t *testing.T) {
	mocker := NewMocker()
	if mocker.info == nil {
		t.Error("expected info to be initialized")
	}
	if mocker.info.Types == nil {
		t.Error("expected Types map to be initialized")
	}
	if mocker.info.Defs == nil {
		t.Error("expected Defs map to be initialized")
	}
	if mocker.info.Uses == nil {
		t.Error("expected Uses map to be initialized")
	}
	if mocker.info.Implicits == nil {
		t.Error("expected Implicits map to be initialized")
	}
	if mocker.info.Selections == nil {
		t.Error("expected Selections map to be initialized")
	}
	if mocker.info.Scopes == nil {
		t.Error("expected Scopes map to be initialized")
	}
	if mocker.info.InitOrder == nil {
		t.Error("expected InitOrder to be initialized")
	}

	analyzer := mocker.Analyzer()
	if analyzer.info != mocker.info {
		t.Error("expected analyzer to use mocker's info")
	}
	if analyzer.all == nil {
		t.Error("expected all map to be initialized")
	}
}
