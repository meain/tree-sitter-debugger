# Tree-Sitter Debugger

A debugging tool for tree-sitter grammars and queries.

## Installation

```bash
go install github.com/meain/tree-sitter-debugger@latest
```

Or clone the repo and build it:

```bash
git clone https://github.com/meain/tree-sitter-debugger.git
cd tree-sitter-debugger
go build
```

## Usage

### Basic Usage

```bash
# Parse a file with a specific language
tree-sitter-debugger --lang <language> filename

# Parse stdin
tree-sitter-debugger --lang <language> < input.txt

# Run a specific query
tree-sitter-debugger --lang <language> --query <query> filename
```

### Available Commands

- List supported languages: `tree-sitter-debugger --list-languages`
- Parse with language: `tree-sitter-debugger --lang <language> [filename]`
- Run query: `tree-sitter-debugger --lang <language> --query <query> [filename]`

## Examples

```bash
# List all supported languages
tree-sitter-debugger --list-languages

# Parse a Go file
tree-sitter-debugger --lang go main.go

# Parse a string
echo "package main" | tree-sitter-debugger --lang go

# Run a query on a file
tree-sitter-debugger --lang go --query "(package_clause (package_identifier) @package)" main.go
```