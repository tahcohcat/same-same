# Cobra CLI Migration

This document describes the migration from standalone commands to a unified Cobra-based CLI.

## Summary

The Same-Same CLI has been refactored to use [Cobra](https://github.com/spf13/cobra), providing:
- **Unified command structure** - Single binary with subcommands
- **Better help text** - Consistent, auto-generated help
- **Shell completion** - Auto-completion support
- **Short flags** - `-n` instead of `--namespace`
- **Extensibility** - Easy to add new commands

## Architecture

### Before (Old Structure)
```
cmd/
├── same-same/
│   └── main.go          # Server with flags
└── ingest/
    └── main.go          # Standalone ingest tool
```

**Usage:**
```bash
./same-same -addr :8080 -debug
./ingest -namespace quotes demo
```

### After (New Structure)
```
cmd/
└── same-same/
    ├── main.go          # Entry point
    └── cmd/
        ├── root.go      # Root command + global flags
        ├── serve.go     # Server subcommand
        └── ingest.go    # Ingest subcommand
```

**Usage:**
```bash
same-same serve -a :8080 -d
same-same ingest -n quotes demo
```

## New Command Structure

### Root Command
```bash
same-same --help
```

Global flags (available to all subcommands):
- `-v, --verbose` - Verbose output
- `-n, --namespace` - Namespace for vectors
- `--dry-run` - Dry run mode
- `--version` - Show version

### Serve Command
```bash
same-same serve [flags]
```

Flags:
- `-a, --addr` - HTTP server address (default ":8080")
- `-d, --debug` - Enable debug logging

Examples:
```bash
# Start on default port
same-same serve

# Custom port with debug
same-same serve -a :9000 -d
```

### Ingest Command
```bash
same-same ingest <source> [flags]
```

Flags:
- `--batch-size` - Batch size for operations (default 100)
- `--benchmark` - Run in benchmark mode
- `-e, --embedder` - Embedder type (local, gemini, huggingface)
- `--id-col` - ID column name (default "id")
- `--max-tokens` - Max tokens per document (default 512)
- `--meta-col` - Metadata column name
- `-o, --output` - Output file for export
- `--sample` - Sample N rows (0 = all)
- `--split` - Dataset split (HF only, default "train")
- `--text-col` - Text column name (default "text")
- `--timeout` - Ingestion timeout (default 30m)

Examples:
```bash
# Built-in dataset
same-same ingest demo

# With namespace and verbose
same-same ingest -n quotes -v demo

# CSV with custom column
same-same ingest --text-col content data.csv

# HuggingFace dataset
same-same ingest hf:imdb --split train --sample 1000

# With specific embedder
same-same ingest -e gemini demo
```

## Migration Guide

### For Users

**Old:**
```bash
# Build two binaries
go build ./cmd/same-same
go build ./cmd/ingest

# Run server
./same-same -addr :8080

# Run ingest
./ingest -namespace quotes demo
```

**New:**
```bash
# Build one binary
go build ./cmd/same-same

# Run server
./same-same serve -a :8080

# Run ingest
./same-same ingest -n quotes demo
```

### For Developers

The old standalone `cmd/ingest/main.go` is now deprecated in favor of the unified CLI. To add new commands:

1. Create a new file in `cmd/same-same/cmd/` (e.g., `export.go`)
2. Define your command:

```go
package cmd

import "github.com/spf13/cobra"

var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export vectors to file",
    Run:   runExport,
}

func init() {
    rootCmd.AddCommand(exportCmd)
    exportCmd.Flags().StringP("output", "o", "", "Output file")
}

func runExport(cmd *cobra.Command, args []string) {
    // Implementation
}
```

3. The command is automatically available:

```bash
same-same export -o data.json
```

## Benefits

### 1. **Unified Binary**
- Single executable instead of multiple tools
- Consistent command structure
- Shared global flags

### 2. **Better UX**
- Intuitive command hierarchy
- Comprehensive help text
- Shell completion support

### 3. **Short Flags**
Before: `--namespace`, `--verbose`, `--embedder`
After: `-n`, `-v`, `-e`

### 4. **Extensibility**
Easy to add new commands:
- `same-same export` - Export data
- `same-same backup` - Backup storage
- `same-same migrate` - Migrate data
- `same-same stats` - Show statistics

### 5. **Professional CLI**
Follows Go CLI best practices used by:
- `kubectl` (Kubernetes)
- `gh` (GitHub CLI)
- `docker` (Docker)
- `hugo` (Hugo static site generator)

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
same-same completion bash > /etc/bash_completion.d/same-same

# Zsh
same-same completion zsh > "${fpath[1]}/_same-same"

# Fish
same-same completion fish > ~/.config/fish/completions/same-same.fish

# PowerShell
same-same completion powershell > same-same.ps1
```

## Backwards Compatibility

The old standalone tools can still be built if needed:

```bash
# Build old ingest tool (deprecated)
go build ./cmd/ingest

# Use as before
./ingest demo
```

However, we recommend migrating to the new unified CLI.

## Testing

All functionality from the old CLI has been preserved:

```bash
# Test server
same-same serve &
curl http://localhost:8080/health

# Test ingest
same-same ingest demo
same-same ingest -n test --dry-run -v quotes
same-same ingest .examples/data/sample.csv
```

## Future Enhancements

Planned additions to the CLI:

1. **Export Command**
   ```bash
   same-same export -o backup.json
   ```

2. **Stats Command**
   ```bash
   same-same stats --namespace quotes
   ```

3. **Backup/Restore Commands**
   ```bash
   same-same backup -o backup/
   same-same restore -i backup/
   ```

4. **Interactive Mode**
   ```bash
   same-same interactive
   > ingest demo
   > search "machine learning"
   ```

## Documentation Updates

All documentation has been updated to reflect the new CLI:
- [README.md](README.md) - Getting Started section
- [INGESTION_GUIDE.md](INGESTION_GUIDE.md) - All examples
- [QUICKSTART.md](.examples/QUICKSTART.md) - Quick reference

## Dependencies

New dependency added:
- `github.com/spf13/cobra` v1.10.1
- `github.com/spf13/pflag` v1.0.9 (cobra dependency)

## Questions?

For issues or questions about the new CLI:
- Check `same-same <command> --help`
- See [INGESTION_GUIDE.md](INGESTION_GUIDE.md)
- Open an issue on GitHub
