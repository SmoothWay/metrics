package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/jingyugao/rowserrcheck/passes/rowserr"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

// OSExit is an analyzer that checks for os.Exit function calls in main package.
var OSExit = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for os.Exit in main package",
	Run:  run,
}

func main() {
	checks := []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		bodyclose.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		rowserr.NewAnalyzer(
			"github.com/jackc/pgx/v5",
		),
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	checks = append(checks, OSExit)

	multichecker.Main(
		checks...,
	)

}

// run performs analysis on the given *analysis.Pass object.
//
// It iterates through each file in the pass, inspects the AST nodes for function calls, and reports if os.Exit is called in the main function.
// Returns nil and nil.
func run(pass *analysis.Pass) (interface{}, error) {
skipGenerated:
	for _, file := range pass.Files {
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				// Skip generated files
				if strings.Contains(c.Text, "DO NOT EDIT") {
					continue skipGenerated
				}
			}
		}
		if file.Name.Name != "main" {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if x, ok := n.(*ast.CallExpr); ok {
				if OSExitChecker(pass, x) && pass.Pkg.Name() == "main" {
					pass.Reportf(x.Pos(), "should not call os.Exit in main function")
				}
			}

			return true
		})
	}

	return nil, nil
}

// OSExitChecker checks if the given CallExpr corresponds to an os.Exit call in the main package.
//
// Parameters: pass *analysis.Pass, x *ast.CallExpr
// Returns: bool
func OSExitChecker(pass *analysis.Pass, x *ast.CallExpr) bool {
	if selector, ok := x.Fun.(*ast.SelectorExpr); ok {
		if id, ok := selector.X.(*ast.Ident); ok && id.Name == "os" && selector.Sel.Name == "Exit" && pass.Pkg.Name() == "main" {
			for _, f := range pass.Files {
				if f.Name.Name == "main" {
					return true
				}
			}
		}
	}

	return false
}
