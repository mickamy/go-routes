package framework

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/mickamy/go-routes/internal/loader"
	"github.com/mickamy/go-routes/internal/route"
)

// netHTTP extracts routes registered through the standard library: the package
// functions http.HandleFunc / http.Handle and the *http.ServeMux methods of the
// same names. Go 1.22 method patterns ("GET /x") yield a verb; legacy patterns
// ("/x") report route.VerbUnknown.
type netHTTP struct{}

func (netHTTP) Name() string { return "net/http" }

func (e netHTTP) Extract(l loader.Loaded) ([]route.Route, error) {
	var routes []route.Route
	for _, pkg := range l.Packages {
		if pkg.TypesInfo == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if r, ok := e.routeFrom(pkg, l, call); ok {
					routes = append(routes, r)
				}

				return true
			})
		}
	}

	return routes, nil
}

func (e netHTTP) routeFrom(pkg *packages.Package, l loader.Loaded, call *ast.CallExpr) (route.Route, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) != 2 {
		return route.Route{}, false
	}

	fn, ok := pkg.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || !isNetHTTPRegistration(fn) {
		return route.Route{}, false
	}

	verb, path := patternFrom(pkg, call.Args[0])
	handler, file, line := resolveHandler(pkg, l, call.Args[1])

	return route.Route{
		Framework: e.Name(),
		Verb:      verb,
		Path:      path,
		Handler:   handler,
		File:      file,
		Line:      line,
	}, true
}

func isNetHTTPRegistration(fn *types.Func) bool {
	if fn.Pkg() == nil || fn.Pkg().Path() != "net/http" {
		return false
	}

	return fn.Name() == "HandleFunc" || fn.Name() == "Handle"
}

func patternFrom(pkg *packages.Package, expr ast.Expr) (verb, path string) {
	raw, ok := stringValue(pkg, expr)
	if !ok {
		return route.VerbUnknown, "<dynamic>"
	}

	raw = strings.TrimSpace(raw)
	if verb, rest, found := strings.Cut(raw, " "); found {
		return strings.TrimSpace(verb), stripHost(strings.TrimSpace(rest))
	}

	return route.VerbUnknown, stripHost(raw)
}

// stripHost drops the optional host component of a net/http pattern, keeping the
// path that begins at the first slash.
func stripHost(s string) string {
	if strings.HasPrefix(s, "/") {
		return s
	}
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[i:]
	}

	return s
}

// stringValue resolves expr to a string, preferring the type checker's constant
// folding (so named constants and constant concatenations like apiBase+"/posts"
// resolve) and falling back to a direct string literal.
func stringValue(pkg *packages.Package, expr ast.Expr) (string, bool) {
	if pkg != nil && pkg.TypesInfo != nil {
		if tv, ok := pkg.TypesInfo.Types[expr]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
			return constant.StringVal(tv.Value), true
		}
	}

	return stringLiteral(expr)
}

func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return v, true
}
