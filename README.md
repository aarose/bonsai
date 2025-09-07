# bonsai
git for LLM convos

## Installation

```bash
go install github.com/aarose/bonsai@latest
```

Or build from source:
```bash
git clone https://github.com/aarose/bonsai.git
cd bonsai
go build -o bai
```

## Usage

```bash
bai              # Run the tool
bai version      # Check version
bai --help       # See help
```

## Dev Notes

### Key Features
- Uses Cobra framework for CLI structure
- Base command "bai" with help and version subcommands
- Modular command structure in the `cmd` package
- Makefile for easy building and installation
- Ready for `go install` to make it globally available

### Development Commands
```bash
go mod download    # Install dependencies
go build -o bai    # Build binary
make build         # Build using Makefile
make install       # Install globally
make test          # Run tests
```

### Test fixture -- generate fake convo

Run with default database (bonsai.db)
Default behavior: Uses ~/.bonsai/bonsai.db (same as CLI tool)
```bash
./scripts/generate_fake_data.sh
```

Or specify custom database path
```bash
./scripts/generate_fake_data.sh path/to/your/database.db
```

Or run the Go script directly
```bash
go run scripts/generate_fake_data.go [database_path]
```

### Visualization

Launch visualization (opens on http://localhost:8080)
```bash
./bai visualize
```

Custom port
```bash
./bai visualize --port 3000
```

Custom database
```
./bai visualize --database ./custom.db
```
