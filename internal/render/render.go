// Package render formats extracted routes for output, either as an aligned
// Rails-routes-style table or as JSON for scripting.
package render

import (
	"fmt"
	"io"
	"sort"

	"github.com/mickamy/go-routes/internal/route"
)

// Format selects the output representation.
type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
)

// ParseFormat validates a --format value.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case Table:
		return Table, nil
	case JSON:
		return JSON, nil
	default:
		return "", fmt.Errorf("unknown format %q (want table or json)", s)
	}
}

// Render writes the routes to w in the requested format.
func Render(w io.Writer, routes []route.Route, format Format) error {
	switch format {
	case JSON:
		return renderJSON(w, routes)
	case Table:
		return renderTable(w, routes)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

// frameworks returns the distinct frameworks present, in stable order.
func frameworks(routes []route.Route) []string {
	seen := map[string]bool{}
	var names []string
	for _, r := range routes {
		if r.Framework != "" && !seen[r.Framework] {
			seen[r.Framework] = true
			names = append(names, r.Framework)
		}
	}
	sort.Strings(names)

	return names
}
