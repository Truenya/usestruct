package usestruct

import (
	"encoding/json"
	"fmt"

	"github.com/Truenya/usestruct/pkg/analyzer"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("usestruct", New)
}

// Config holds the configuration for the usestruct analyzer
type Config struct {
	// MinRequiredParams defines the minimum number of parameters required for analysis
	MinRequiredParams int `json:"min_required_params"`
	// MaxRecursionDepth defines the maximum recursion depth when analyzing call chains
	MaxRecursionDepth int `json:"max_recursion_depth"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		MinRequiredParams: 2,
		MaxRecursionDepth: 10,
	}
}

type PluginUsestructModule struct {
	config Config
}

func New(settings any) (register.LinterPlugin, error) {
	config := DefaultConfig()

	if settings == nil {
		return PluginUsestructModule{config: config}, nil
	}

	// Try to parse settings as JSON
	settingsMap, ok := settings.(map[string]any)
	if !ok {
		return PluginUsestructModule{config: config}, nil
	}

	// Convert map to JSON and then to our config struct
	jsonBytes, err := json.Marshal(settingsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	var parsedConfig Config
	if err := json.Unmarshal(jsonBytes, &parsedConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// Use parsed values if they are valid (> 0), otherwise keep defaults
	if parsedConfig.MinRequiredParams > 0 {
		config.MinRequiredParams = parsedConfig.MinRequiredParams
	}
	if parsedConfig.MaxRecursionDepth > 0 {
		config.MaxRecursionDepth = parsedConfig.MaxRecursionDepth
	}

	return PluginUsestructModule{config: config}, nil
}

func (f PluginUsestructModule) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.AnalyzerWithConfig(f.config.MinRequiredParams, f.config.MaxRecursionDepth),
	}, nil
}

func (f PluginUsestructModule) GetLoadMode() string {
	return register.LoadModeSyntax
}
