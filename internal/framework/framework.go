// Package framework holds the per-framework AST visitors that recognize route
// registration calls (e.g., e.GET("/x", h)) and resolve their handlers to a
// pkg.Func symbol with a file:line, plus the shared helpers they build on.
package framework

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/mickamy/go-routes/internal/loader"
	"github.com/mickamy/go-routes/internal/route"
)

// Extractor recognizes the route registrations of a single web framework.
type Extractor interface {
	// Name is the framework identifier shown in output and accepted by
	// --framework (e.g., "echo", "net/http").
	Name() string
	// Extract walks the loaded packages and returns every route it can resolve.
	Extract(l loader.Loaded) ([]route.Route, error)
}

func registry() []Extractor {
	return []Extractor{
		netHTTP{},
	}
}

// Extract runs the framework visitors over the loaded packages. When only is
// non-empty, just that framework's extractor runs; otherwise every known
// extractor runs and routes from all of them are merged.
func Extract(l loader.Loaded, only string) ([]route.Route, error) {
	extractors := registry()
	if only != "" {
		matched := false
		for _, e := range extractors {
			if e.Name() == only {
				extractors = []Extractor{e}
				matched = true

				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("unknown framework %q (known: %s)", only, knownNames())
		}
	}

	var routes []route.Route
	for _, e := range extractors {
		rs, err := e.Extract(l)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		routes = append(routes, rs...)
	}

	return routes, nil
}

func knownNames() string {
	reg := registry()
	names := make([]string, 0, len(reg))
	for _, e := range reg {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	return strings.Join(names, ", ")
}

// resolveHandler turns the handler argument of a route registration into a
// display symbol and its declaration site. Named functions and methods resolve
// through type info; function literals fall back to their source position.
func resolveHandler(pkg *packages.Package, l loader.Loaded, expr ast.Expr) (handler, file string, line int) {
	expr = unwrap(expr)

	if lit, ok := expr.(*ast.FuncLit); ok {
		pos := l.Fset.Position(lit.Pos())

		return "func", l.RelPath(pos.Filename), pos.Line
	}

	obj := referencedObject(pkg, expr)
	fn, ok := obj.(*types.Func)
	if !ok {
		pos := l.Fset.Position(expr.Pos())

		return "?", l.RelPath(pos.Filename), pos.Line
	}

	pos := l.Fset.Position(fn.Pos())

	return funcSymbol(fn), l.RelPath(pos.Filename), pos.Line
}

// unwrap strips a single http.HandlerFunc(...) style conversion so the inner
// handler expression can be resolved directly.
func unwrap(expr ast.Expr) ast.Expr {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return expr
	}

	switch call.Fun.(type) {
	case *ast.SelectorExpr, *ast.Ident:
		return call.Args[0]
	default:
		return expr
	}
}

func referencedObject(pkg *packages.Package, expr ast.Expr) types.Object {
	if pkg == nil || pkg.TypesInfo == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return pkg.TypesInfo.Uses[e]
	case *ast.SelectorExpr:
		return pkg.TypesInfo.Uses[e.Sel]
	default:
		return nil
	}
}

// funcSymbol renders a *types.Func as "pkg.Func" or "pkg.Type.Method".
func funcSymbol(fn *types.Func) string {
	pkgName := ""
	if fn.Pkg() != nil {
		pkgName = fn.Pkg().Name()
	}

	if recv := receiverTypeName(fn); recv != "" {
		if pkgName != "" {
			return pkgName + "." + recv + "." + fn.Name()
		}

		return recv + "." + fn.Name()
	}

	if pkgName != "" {
		return pkgName + "." + fn.Name()
	}

	return fn.Name()
}

func receiverTypeName(fn *types.Func) string {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return ""
	}

	t := sig.Recv().Type()
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if named, ok := t.(*types.Named); ok {
		return named.Obj().Name()
	}

	return ""
}
