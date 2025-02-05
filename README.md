

# usestruct

A Go linter plugin that suggests using structs when passing multiple parameters through function call chains.

## Description

`usestruct` analyzes your Go codebase and identifies patterns where it would be more efficient or cleaner to use a struct to group function parameters. This helps improve code clarity, maintainability, and readability by suggesting better ways to handle parameter passing in complex function call chains.

Instead of:

```go
func configureServer(port int, address string, timeout time.Duration) {
    // function implementation
}

configureServer(8080, "localhost", 30*time.Second)
```

`usestruct` suggests refactoring to use a struct:

```go
type ServerConfig struct {
    Port     int
    Address  string
    Timeout  time.Duration
}

func configureServer(config ServerConfig) {
    // function implementation
}

config := ServerConfig{
    Port:    8080,
    Address: "localhost",
    Timeout: 30 * time.Second,
}
configureServer(config)
```

## Features

- Analyzes function call chains to identify opportunities for struct usage
- Suggests creating struct types to group related parameters
- Improves code readability and maintainability
- Highly configurable with customizable threshold and depth parameters

## Installation

To install the `usestruct` linter:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint install-usestruct
```

Or add it directly to your `.golangci.yml` configuration file:

```yaml
linters:
  enable:
    - usestruct
```

## Usage

To use `usestruct`, simply run the linter on your codebase with GolangCI-Lint:

```sh
golangci-lint run
```

The tool will analyze your function calls and suggest struct creation when it detects patterns of multiple parameters being passed through function chains.

## Configuration

You can customize the behavior of `usestruct` by adding configuration options to your `.golangci.yml` file:

```yaml
linters-settings:
  custom:
    usestruct:
      settings:
        min_required_params: 3    # Minimum number of parameters to trigger analysis (default: 2)
        max_recursion_depth: 15   # Maximum recursion depth for call chain analysis (default: 10)
```

### Configuration Options

- `min_required_params` (default: 2): The minimum number of parameters a function must have before the analyzer considers suggesting a struct. Functions with fewer parameters will be ignored.

- `max_recursion_depth` (default: 10): The maximum depth the analyzer will traverse when following function call chains. This prevents infinite recursion and controls analysis performance.

## How It Works

`usostruct` analyzes Go functions that accept two or more parameters, tracking how these parameters are used in nested function calls. When it identifies a chain of function calls where the same set of parameters is repeatedly passed along, it suggests creating a struct to group those parameters together.

The analyzer has built-in limits:
- Minimum required parameters: 2
- Maximum recursion depth: 10

These thresholds can be adjusted as needed based on your project's requirements.

