# dicebear-shim

dicebear-shim CLI tool.

<!--doc-start-->
## Installation

### From Source

```sh
git clone https://github.com/kettleofketchup/diceavatar-shim.git
cd diceavatar-shim
./dev  # Bootstrap environment
just build
```

### Docker

```sh
docker pull ghcr.io/kettleofketchup/diceavatar-shim/dicebear-shim:latest
docker run --rm ghcr.io/kettleofketchup/diceavatar-shim/dicebear-shim:latest --help
```

## Usage

```sh
# Show help
./bin/dicebear-shim --help

# Show version
./bin/dicebear-shim version

# Use with config file
./bin/dicebear-shim --config ./config/dicebear-shim.yaml
```

## Development

### Prerequisites

- Go 1.23+
- golangci-lint
- [just](https://github.com/casey/just) (auto-installed by `./dev`)
- uv (for documentation)

### Quick Start

```sh
./dev  # Bootstrap environment, install just if needed
```

### Build

```sh
just build          # Build binary
just test           # Run tests
just lint           # Run linter
just release::all   # Build for all platforms
```

### Documentation

```sh
just docs::serve    # Start dev server at localhost:8000
just docs::build    # Build static documentation
```

### Docker

```sh
just docker::build  # Build Docker image
just docker::push   # Push to registry
```

### Copier Template

This project was generated from a copier template. To update the project from the latest template or change your copier answers:

```sh
just copier::update     # Update from template (re-prompts for answers)
just copier::update-auto # Update with current answers (no prompts)
just copier::diff       # Preview changes without applying
just copier::recopy     # Full re-copy (after major template changes)
just copier::answers    # Show current template answers
```

To change your copier answers (e.g. project name, description, options), run `just copier::update` — it will re-prompt for each answer, letting you modify them.
<!--doc-end-->

## License

[Add your license here]
