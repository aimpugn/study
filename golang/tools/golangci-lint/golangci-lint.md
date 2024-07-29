# [golangci-lint](https://golangci-lint.run/)

- [golangci-lint](#golangci-lint)
    - [installation](#installation)
    - [`--help`](#--help)
    - [`run`](#run)
        - [Usage](#usage)

## installation

```bash
brew install golangci-lint
```

## `--help`

Smart, fast linters runner.

Usage:
  golangci-lint [flags]
  golangci-lint [command]

Available Commands:
  cache       Cache control and information
  completion  Generate the autocompletion script for the specified shell
  config      Config
  help        Help
  linters     List current linters configuration
  run         Run the linters
  version     Version

Flags:
      --color string              Use color when printing; can be 'always', 'auto', or 'never' (default "auto")
  -j, --concurrency int           Concurrency (default NumCPU) (default 10)
      --cpu-profile-path string   Path to CPU profile output file
  -h, --help                      help for golangci-lint
      --mem-profile-path string   Path to memory profile output file
      --trace-path string         Path to trace output file
  -v, --verbose                   Verbose output
      --version                   Print version

Use "golangci-lint [command] --help" for more information about a command.

## `run`

Run the linters

### Usage

```bash
golangci-lint run [flags]
```

```bash
golangci-lint run -p 'bugs,unused,error,format,complexity,performance,import,comment' ./...
```

```text
start 2024-02-06 09:00:00.000 +09:00 KST
end 2024-02-11 00:00:00.000 +09:00 KST
exhausted 2024-02-07 09:49:22.000 +09:00 KST
```
