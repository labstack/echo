package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func badArgument(fileset *token.FileSet, p token.Pos) {
	pos := fileset.Position(p)
	filename := pos.Filename
	base, err := os.Getwd()
	if err == nil {
		rpath, perr := filepath.Rel(base, pos.Filename)
		if perr == nil {
			filename = rpath
		}
	}
	msg := fmt.Sprintf("%s:%d: Error: found call to rice.FindBox, "+
		"but argument must be a string literal.\n", filename, pos.Line)
	fmt.Println(msg)
	os.Exit(1)
}

func findBoxes(pkg *build.Package) map[string]bool {
	// create map of boxes to embed
	var boxMap = make(map[string]bool)

	// create one list of files for this package
	filenames := make([]string, 0, len(pkg.GoFiles)+len(pkg.CgoFiles))
	filenames = append(filenames, pkg.GoFiles...)
	filenames = append(filenames, pkg.CgoFiles...)

	// loop over files, search for rice.FindBox(..) calls
	for _, filename := range filenames {
		// find full filepath
		fullpath := filepath.Join(pkg.Dir, filename)
		if strings.HasSuffix(filename, "rice-box.go") {
			// Ignore *.rice-box.go files
			verbosef("skipping file %q\n", fullpath)
			continue
		}
		verbosef("scanning file %q\n", fullpath)

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fullpath, nil, 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var riceIsImported bool
		ricePkgName := "rice"
		for _, imp := range f.Imports {
			if strings.HasSuffix(imp.Path.Value, "go.rice\"") {
				if imp.Name != nil {
					ricePkgName = imp.Name.Name
				}
				riceIsImported = true
				break
			}
		}
		if !riceIsImported {
			// Rice wasn't imported, so we won't find a box.
			continue
		}
		if ricePkgName == "_" {
			// Rice pkg is unnamed, so we won't find a box.
			continue
		}

		// Inspect AST, looking for calls to (Must)?FindBox.
		// First parameter of the func must be a basic literal.
		// Identifiers won't be resolved.
		var nextIdentIsBoxFunc bool
		var nextBasicLitParamIsBoxName bool
		var boxCall token.Pos
		var variableToRemember string
		var validVariablesForBoxes map[string]bool = make(map[string]bool)

		ast.Inspect(f, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			switch x := node.(type) {
			// this case fixes the var := func() style assignments, not assignments to vars declared separately from the assignment.
			case *ast.AssignStmt:
				var assign = node.(*ast.AssignStmt)
				name, found := assign.Lhs[0].(*ast.Ident)
				if found {
					variableToRemember = name.Name
					composite, first := assign.Rhs[0].(*ast.CompositeLit)
					if first {
						riceSelector, second := composite.Type.(*ast.SelectorExpr)

						if second {
							callCorrect := riceSelector.Sel.Name == "Config"
							packageName, third := riceSelector.X.(*ast.Ident)

							if third && callCorrect && packageName.Name == ricePkgName {
								validVariablesForBoxes[name.Name] = true
								verbosef("\tfound variable, saving to scan for boxes: %q\n", name.Name)
							}
						}
					}
				}
			case *ast.Ident:
				if nextIdentIsBoxFunc || ricePkgName == "." {
					nextIdentIsBoxFunc = false
					if x.Name == "FindBox" || x.Name == "MustFindBox" {
						nextBasicLitParamIsBoxName = true
						boxCall = x.Pos()
					}
				} else {
					if x.Name == ricePkgName || validVariablesForBoxes[x.Name] {
						nextIdentIsBoxFunc = true
					}
				}
			case *ast.BasicLit:
				if nextBasicLitParamIsBoxName {
					if x.Kind == token.STRING {
						nextBasicLitParamIsBoxName = false
						// trim "" or ``
						name := x.Value[1 : len(x.Value)-1]
						boxMap[name] = true
						verbosef("\tfound box %q\n", name)
					} else {
						badArgument(fset, boxCall)
					}
				}

			default:
				if nextIdentIsBoxFunc {
					nextIdentIsBoxFunc = false
				}
				if nextBasicLitParamIsBoxName {
					badArgument(fset, boxCall)
				}
			}
			return true
		})
	}

	return boxMap
}
