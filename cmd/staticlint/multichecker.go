// Analyzer set of analyzers for static analysis
//
// Contains:
// contains all analyzers of golang.org/x/tools/go/analysis
//
// all analyzers of https://pkg.go.dev/honnef.co/go/tools@v0.4.2/staticcheck and https://pkg.go.dev/honnef.co/go/tools@v0.4.2/stylecheck
//
// OsExitAnalyzer - checks for a call os.Exit() in main method.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// OsExitAnalyzer is a custom os.Exit analyzer
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "os_exit_analyzer",
	Doc:  "check os.Exit command existence",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if p, ok := node.(*ast.Package); ok {
				if p.Name != "main" {
					return true
				}
			}
			if f, ok := node.(*ast.FuncDecl); ok {
				if f.Name.Name == "main" {
					for _, stmt := range f.Body.List {
						if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
							if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
								if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
									if c, ok := selExpr.X.(*ast.Ident); ok {
										if c.Name == "os" && selExpr.Sel.Name == "Exit" {
											pass.Reportf(c.NamePos, "found os.Exit in main func of package main")
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	mychecks = append(
		mychecks,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		nilfunc.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		OsExitAnalyzer,
	)

	multichecker.Main(
		mychecks...,
	)
}
