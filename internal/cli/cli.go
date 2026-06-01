package cli

import (
	"fmt"
	"io"

	"github.com/mickamy/go-routes/internal/exit"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		PrintUsage(stderr)

		return exit.Usage
	}

	fmt.Fprintf(stderr, "go-routes: not yet implemented.\n")

	return exit.NotImplemented
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
	fmt.Fprintln(w, "  --no-color           Disable ANSI color (CI / pipe friendly)")
	fmt.Fprintln(w, "  --version, -v        Print go-routes version")
	fmt.Fprintln(w, "  --help, -h           Show this help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supports echo, gin, chi, gorilla/mux, and net/http (1.22+).")
	fmt.Fprintln(w, "More: https://github.com/mickamy/go-routes")
}
