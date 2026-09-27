// SPDX-License-Identifier: MIT

// config-fields emits source-backed middleware config fields for the website.
// It does not claim to determine runtime defaults or behavior.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type field struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Deprecated bool   `json:"deprecated,omitempty"`
	Line       int    `json:"line"`
}

type config struct {
	Name   string  `json:"name"`
	File   string  `json:"file"`
	Line   int     `json:"line"`
	Fields []field `json:"fields"`
}

type manifest struct {
	Module   string   `json:"module"`
	Revision string   `json:"revision,omitempty"`
	Configs  []config `json:"configs"`
}

func main() {
	root := flag.String("root", "../..", "Echo repository root, relative to the current directory")
	revision := flag.String("revision", "", "exact Echo commit or tag used for this build")
	flag.Parse()

	result, err := extract(*root, *revision)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func extract(root, revision string) (manifest, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return manifest{}, err
	}
	fs := token.NewFileSet()
	packageDir := filepath.Join(root, "middleware")
	packages, err := parser.ParseDir(fs, packageDir, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return manifest{}, err
	}
	pkg, ok := packages["middleware"]
	if !ok {
		return manifest{}, fmt.Errorf("middleware package not found in %s", packageDir)
	}

	result := manifest{Module: "github.com/labstack/echo/v5", Revision: revision, Configs: []config{}}
	for filename, source := range pkg.Files {
		for _, declaration := range source.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.TYPE {
				continue
			}
			for _, item := range group.Specs {
				typeSpec := item.(*ast.TypeSpec)
				structure, ok := typeSpec.Type.(*ast.StructType)
				if !ok || !ast.IsExported(typeSpec.Name.Name) || !strings.HasSuffix(typeSpec.Name.Name, "Config") {
					continue
				}
				entry := config{
					Name:   typeSpec.Name.Name,
					File:   filepath.ToSlash(strings.TrimPrefix(filename, root+string(filepath.Separator))),
					Line:   fs.Position(typeSpec.Pos()).Line,
					Fields: []field{},
				}
				for _, sourceField := range structure.Fields.List {
					var rendered bytes.Buffer
					if err := format.Node(&rendered, fs, sourceField.Type); err != nil {
						return manifest{}, err
					}
					fieldDoc := comment(sourceField.Doc)
					for _, name := range sourceField.Names {
						if !ast.IsExported(name.Name) {
							continue
						}
						entry.Fields = append(entry.Fields, field{
							Name:       name.Name,
							Type:       rendered.String(),
							Deprecated: strings.Contains(fieldDoc, "Deprecated:"),
							Line:       fs.Position(name.Pos()).Line,
						})
					}
				}
				result.Configs = append(result.Configs, entry)
			}
		}
	}
	sort.Slice(result.Configs, func(i, j int) bool { return result.Configs[i].Name < result.Configs[j].Name })
	return result, nil
}

func comment(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return strings.TrimSpace(group.Text())
}
