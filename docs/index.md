# dicebear-shim

Welcome to the dicebear-shim documentation.

## Quick Start

### Installation

```sh
# Bootstrap environment (installs just if needed)
./dev

# Build from source
just build

# Run
./bin/dicebear-shim --help
```

### Version

```sh
./bin/dicebear-shim version
```

## Development

### Build

```sh
just build          # Build binary
just test           # Run tests
just lint           # Run linter
```

### Documentation

```sh
just docs::serve    # Start dev server
just docs::build    # Build static site
```

## Project Structure

```
dicebear-shim/
├── src/dicebear-shim/    # Go source code
│   ├── cmd/                # CLI commands
│   ├── internal/           # Private packages
│   └── version/            # Version info
├── docs/                   # Documentation
├── just/                   # Build recipes (modules)
└── docker/                 # Docker configuration
```


{% include-markdown "../README.md" start="<!--doc-start-->" end="<!--doc-end-->" %}

