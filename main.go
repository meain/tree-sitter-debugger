package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_css "github.com/tree-sitter/tree-sitter-css/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

var supportedLanguages = map[string]*sitter.Language{
	"bash":       sitter.NewLanguage(tree_sitter_bash.Language()),
	"c":          sitter.NewLanguage(tree_sitter_c.Language()),
	"cpp":        sitter.NewLanguage(tree_sitter_cpp.Language()),
	"css":        sitter.NewLanguage(tree_sitter_css.Language()),
	"go":         sitter.NewLanguage(tree_sitter_go.Language()),
	"html":       sitter.NewLanguage(tree_sitter_html.Language()),
	"java":       sitter.NewLanguage(tree_sitter_java.Language()),
	"javascript": sitter.NewLanguage(tree_sitter_javascript.Language()),
	"js":         sitter.NewLanguage(tree_sitter_javascript.Language()),
	"php":        sitter.NewLanguage(tree_sitter_php.LanguagePHP()),
	"python":     sitter.NewLanguage(tree_sitter_python.Language()),
	"py":         sitter.NewLanguage(tree_sitter_python.Language()),
	"ruby":       sitter.NewLanguage(tree_sitter_ruby.Language()),
	"rust":       sitter.NewLanguage(tree_sitter_rust.Language()),
	"typescript": sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()),
	"ts":         sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()),
	"tsx":        sitter.NewLanguage(tree_sitter_typescript.LanguageTSX()),
}

func main() {
	var (
		lang          = flag.String("lang", "", "Language to parse (required)")
		query         = flag.String("query", "", "Tree-sitter query to execute")
		listLanguages = flag.Bool("list-languages", false, "List all supported languages")
	)
	flag.Parse()

	// Check if we just need to list languages
	if *listLanguages {
		fmt.Println("Supported languages:")
		for _, lang := range getSupportedLanguages() {
			fmt.Println(" -", lang)
		}
		os.Exit(0)
	}

	if *lang == "" {
		fmt.Fprintf(os.Stderr, "Error: --lang is required\n")
		fmt.Fprintf(os.Stderr, "Use --list-languages to see all supported languages\n")
		os.Exit(1)
	}

	language, ok := supportedLanguages[*lang]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unsupported language '%s'\n", *lang)
		fmt.Fprintf(os.Stderr, "Supported languages: %s\n", strings.Join(getSupportedLanguages(), ", "))
		os.Exit(1)
	}

	// Handle positional arguments for filename
	args := flag.Args()

	// Read input
	var input []byte
	var err error
	if len(args) > 0 {
		input, err = os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse the code
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(language)

	tree := parser.Parse(input, nil)
	defer tree.Close()

	if *query != "" {
		// Execute query
		if err := executeQuery(tree, language, input, *query); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Print the tree
		printTree(tree.RootNode(), input, 0)
	}
}

func executeQuery(
	tree *sitter.Tree,
	language *sitter.Language,
	source []byte,
	queryStr string,
) error {
	q, err := sitter.NewQuery(language, queryStr)
	if err != nil {
		return fmt.Errorf("invalid query: %v", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	matches := qc.Matches(q, tree.RootNode(), source)

	matchCount := 0
	for {
		match := matches.Next()
		if match == nil {
			break
		}

		matchCount++
		if matchCount > 1 {
			fmt.Println()
		}

		for _, capture := range match.Captures {
			captureName := q.CaptureNames()[capture.Index]
			node := capture.Node

			startPoint := node.StartPosition()
			endPoint := node.EndPosition()

			fmt.Printf("@%s\n", captureName)
			fmt.Printf("start: %d:%d\n", startPoint.Row+1, startPoint.Column)
			fmt.Printf("end: %d:%d\n", endPoint.Row+1, endPoint.Column)
			fmt.Printf("content:\n")

			// Extract the content
			content := source[node.StartByte():node.EndByte()]

			// Print each line with some indentation for readability
			scanner := bufio.NewScanner(strings.NewReader(string(content)))
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
			fmt.Println()
		}
	}

	if matchCount == 0 {
		fmt.Println("No matches found")
	}

	return nil
}

func printTree(node *sitter.Node, source []byte, depth int) {
	indent := strings.Repeat("  ", depth)
	nodeType := node.Kind()

	if node.IsNamed() {
		if node.ChildCount() == 0 {
			// Leaf node - show content
			content := source[node.StartByte():node.EndByte()]
			// Escape newlines and tabs for display
			displayContent := strings.ReplaceAll(string(content), "\n", "\\n")
			displayContent = strings.ReplaceAll(displayContent, "\t", "\\t")
			if len(displayContent) > 50 {
				displayContent = displayContent[:47] + "..."
			}
			fmt.Printf("%s(%s \"%s\")\n", indent, nodeType, displayContent)
		} else {
			fmt.Printf("%s(%s\n", indent, nodeType)
		}
	} else {
		// Anonymous node
		content := source[node.StartByte():node.EndByte()]
		displayContent := strings.ReplaceAll(string(content), "\n", "\\n")
		displayContent = strings.ReplaceAll(displayContent, "\t", "\\t")
		fmt.Printf("%s\"%s\"\n", indent, displayContent)
	}

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)
		printTree(child, source, depth+1)
	}

	if node.IsNamed() && node.ChildCount() > 0 {
		fmt.Printf("%s)\n", indent)
	}
}

func getSupportedLanguages() []string {
	langs := make([]string, 0, len(supportedLanguages))
	for lang := range supportedLanguages {
		langs = append(langs, lang)
	}
	// Sort languages alphabetically for consistent output
	sort.Strings(langs)
	return langs
}
