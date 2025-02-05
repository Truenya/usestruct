package usestruct

import (
	"github.com/Truenya/usestruct/pkg/analyzer"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("usestruct", New)
}

type PluginUsestructModule struct{}

func New(settings any) (register.LinterPlugin, error) {
	return PluginUsestructModule{}, nil
}

func (f PluginUsestructModule) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.Analyzer(),
	}, nil
}

func (f PluginUsestructModule) GetLoadMode() string {
	return register.LoadModeSyntax
}
