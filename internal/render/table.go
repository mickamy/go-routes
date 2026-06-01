package render

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/mickamy/go-routes/internal/route"
)

func renderTable(w io.Writer, routes []route.Route) error {
	if len(routes) == 0 {
		fmt.Fprintln(w, "no routes found")

		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "VERB\tPATH\tHANDLER\tFILE:LINE")
	for _, r := range routes {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s:%d\n", r.Verb, r.Path, r.Handler, r.File, r.Line)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flushing table: %w", err)
	}

	fmt.Fprintf(w, "\n%s\n", summary(routes))

	return nil
}

// summary renders the footer line, e.g., "8 routes in 2 file(s) · echo".
func summary(routes []route.Route) string {
	files := map[string]bool{}
	for _, r := range routes {
		files[r.File] = true
	}

	line := fmt.Sprintf("%s in %s", plural(len(routes), "route"), plural(len(files), "file"))
	if fws := frameworks(routes); len(fws) > 0 {
		line += " · " + strings.Join(fws, ", ")
	}

	return line
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}

	return fmt.Sprintf("%d %ss", n, noun)
}
