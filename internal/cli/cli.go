package cli

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.com/mickamy/go-routes/internal/exit"
	"github.com/mickamy/go-routes/internal/framework"
	"github.com/mickamy/go-routes/internal/loader"
	"github.com/mickamy/go-routes/internal/render"
	"github.com/mickamy/go-routes/internal/route"
)

func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("go-routes", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { PrintUsage(stderr) }

	var (
		fwName  = fs.String("framework", "", "override framework auto-detection")
		format  = fs.String("format", "table", "output format: table, json")
		sortKey = fs.String("sort", "", "sort by: path, file, handler (default file order)")
		filter  = fs.String("filter", "", "keep routes whose path or handler matches the regex")
	)

	if err := fs.Parse(args); err != nil {
		return exit.Usage
	}

	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}

	format2, err := render.ParseFormat(*format)
	if err != nil {
		fmt.Fprintf(stderr, "go-routes: %v\n", err)

		return exit.Usage
	}

	var filterRE *regexp.Regexp
	if *filter != "" {
		filterRE, err = regexp.Compile(*filter)
		if err != nil {
			fmt.Fprintf(stderr, "go-routes: invalid --filter regex: %v\n", err)

			return exit.Usage
		}
	}

	loaded, err := loader.Load(dir)
	if err != nil {
		fmt.Fprintf(stderr, "go-routes: %v\n", err)

		return exit.Error
	}

	routes, err := framework.Extract(loaded, *fwName)
	if err != nil {
		fmt.Fprintf(stderr, "go-routes: %v\n", err)

		return exit.Error
	}

	routes = applyFilter(routes, filterRE)
	if err := sortRoutes(routes, *sortKey); err != nil {
		fmt.Fprintf(stderr, "go-routes: %v\n", err)

		return exit.Usage
	}

	if err := render.Render(stdout, routes, format2); err != nil {
		fmt.Fprintf(stderr, "go-routes: %v\n", err)

		return exit.Error
	}

	return exit.OK
}

func applyFilter(routes []route.Route, re *regexp.Regexp) []route.Route {
	if re == nil {
		return routes
	}

	kept := routes[:0]
	for _, r := range routes {
		if re.MatchString(r.Path) || re.MatchString(r.Handler) {
			kept = append(kept, r)
		}
	}

	return kept
}

func sortRoutes(routes []route.Route, key string) error {
	less, err := lessFunc(key)
	if err != nil {
		return err
	}

	sort.SliceStable(routes, func(i, j int) bool {
		return less(routes[i], routes[j])
	})

	return nil
}

func lessFunc(key string) (func(a, b route.Route) bool, error) {
	switch key {
	case "", "file":
		return func(a, b route.Route) bool {
			if a.File != b.File {
				return a.File < b.File
			}

			return a.Line < b.Line
		}, nil
	case "path":
		return func(a, b route.Route) bool {
			if a.Path != b.Path {
				return a.Path < b.Path
			}

			return a.Verb < b.Verb
		}, nil
	case "handler":
		return func(a, b route.Route) bool { return a.Handler < b.Handler }, nil
	default:
		return nil, fmt.Errorf("unknown --sort key %q (want path, file, or handler)", key)
	}
}

func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, "go-routes — bin/rails routes for any Go web framework.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "USAGE:")
	fmt.Fprintln(w, "  go-routes [path] [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Statically analyzes Go source and prints every route the app exposes,")
	fmt.Fprintln(w, "  with its HTTP verb, path, handler symbol, and file:line. No build or run")
	fmt.Fprintln(w, "  required — it reads the AST. Defaults to the current directory.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "EXAMPLES:")
	fmt.Fprintln(w, "  go-routes .")
	fmt.Fprintln(w, "  go-routes ./cmd/server")
	fmt.Fprintln(w, "  go-routes --framework echo .")
	fmt.Fprintln(w, "  go-routes --format json .")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "FLAGS:")
	fmt.Fprintln(w, "  --framework <name>   Override framework auto-detection")
	fmt.Fprintln(w, "  --format <fmt>       Output format: table, json (default table)")
	fmt.Fprintln(w, "  --sort <key>         Sort by: path, file, handler (default file order)")
	fmt.Fprintln(w, "  --filter <regex>     Keep routes whose path or handler matches")
	fmt.Fprintln(w, "  --version, -v        Print go-routes version")
	fmt.Fprintln(w, "  --help, -h           Show this help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supports net/http (1.22+). echo, gin, chi, and gorilla/mux are coming.")
	fmt.Fprintln(w, "More: https://github.com/mickamy/go-routes")
}
