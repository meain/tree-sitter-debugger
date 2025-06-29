package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

var supportedLanguages = map[string]*sitter.Language{
	"bash":       bash.GetLanguage(),
	"c":          c.GetLanguage(),
	"cpp":        cpp.GetLanguage(),
	"css":        css.GetLanguage(),
	"go":         golang.GetLanguage(),
	"html":       html.GetLanguage(),
	"java":       java.GetLanguage(),
	"javascript": typescript.GetLanguage(),
	"js":         typescript.GetLanguage(),
	"php":        php.GetLanguage(),
	"python":     python.GetLanguage(),
	"py":         python.GetLanguage(),
	"ruby":       ruby.GetLanguage(),
	"rust":       rust.GetLanguage(),
	"tsx":        tsx.GetLanguage(),
	"typescript": typescript.GetLanguage(),
	"ts":         typescript.GetLanguage(),
	"yaml":       yaml.GetLanguage(),
	"yml":        yaml.GetLanguage(),
}

func main() {
	var (
		lang  = flag.String("lang", "", "Language to parse (required)")
		query = flag.String("query", "", "Tree-sitter query to execute")
	)
	flag.Parse()

	if *lang == "" {
		fmt.Fprintf(os.Stderr, "Error: --lang is required\n")
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
	parser.SetLanguage(language)

	tree, err := parser.ParseCtx(context.Background(), nil, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing code: %v\n", err)
		os.Exit(1)
	}
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
	q, err := sitter.NewQuery([]byte(queryStr), language)
	if err != nil {
		return fmt.Errorf("invalid query: %v", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, tree.RootNode())

	matchCount := 0
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		matchCount++
		if matchCount > 1 {
			fmt.Println()
		}

		for _, c := range m.Captures {
			captureName := q.CaptureNameForId(c.Index)
			node := c.Node

			startPoint := node.StartPoint()
			endPoint := node.EndPoint()

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
	nodeType := node.Type()

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

	for i := 0; i < int(node.ChildCount()); i++ {
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
	return langs
}
