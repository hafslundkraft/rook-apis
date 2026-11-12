// Command check-upstream checks that every type declared in a copy's package directory, doc comment
// included, is byte for byte the type of the same name in an upstream package directory.
//
// Usage: check-upstream <copy dir> <upstream dir>
package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: check-upstream <copy dir> <upstream dir>")
		os.Exit(2)
	}
	copied, err := typeDecls(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	upstream, err := typeDecls(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	failed := false
	for _, name := range slices.Sorted(maps.Keys(copied)) {
		theirs, ok := upstream[name]
		if !ok {
			fmt.Printf("%s: not declared upstream\n", name)
			failed = true
			continue
		}
		if line, ours, upstreamLine, differ := firstDifference(copied[name], theirs); differ {
			fmt.Printf("%s: line %d of the declaration differs\n  copy:     %s\n  upstream: %s\n",
				name, line, ours, upstreamLine)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
	fmt.Printf("%d types match upstream\n", len(copied))
}

// typeDecls returns the source of each type declaration in the Go files of dir, doc comment
// included, keyed by type name. Generated and test files are skipped.
func typeDecls(dir string) (map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("list Go files in %s: %w", dir, err)
	}
	decls := map[string]string{}
	fset := token.NewFileSet()
	for _, path := range paths {
		name := filepath.Base(path)
		if strings.HasPrefix(name, "zz_generated") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			start := gen.Pos()
			if gen.Doc != nil {
				start = gen.Doc.Pos()
			}
			text := string(src[fset.Position(start).Offset:fset.Position(gen.End()).Offset])
			for _, spec := range gen.Specs {
				decls[spec.(*ast.TypeSpec).Name.Name] = text
			}
		}
	}
	if len(decls) == 0 {
		return nil, errors.New("no type declarations in " + dir)
	}
	return decls, nil
}

// firstDifference returns the first line, counted from 1, at which a and b differ.
func firstDifference(a, b string) (line int, lineA, lineB string, differ bool) {
	linesA, linesB := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := range max(len(linesA), len(linesB)) {
		var x, y string
		if i < len(linesA) {
			x = linesA[i]
		}
		if i < len(linesB) {
			y = linesB[i]
		}
		if x != y {
			return i + 1, x, y, true
		}
	}
	return 0, "", "", false
}
