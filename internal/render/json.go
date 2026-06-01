package render

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mickamy/go-routes/internal/route"
)

type jsonOutput struct {
	Framework string        `json:"framework,omitempty"`
	Routes    []route.Route `json:"routes"`
}

func renderJSON(w io.Writer, routes []route.Route) error {
	out := jsonOutput{Routes: routes}

	// With a single framework, hoist it to a top-level field and drop the
	// redundant per-route tag. With several, keep it on each route instead.
	if fws := frameworks(routes); len(fws) == 1 {
		out.Framework = fws[0]
		stripped := make([]route.Route, len(routes))
		for i, r := range routes {
			r.Framework = ""
			stripped[i] = r
		}
		out.Routes = stripped
	}

	if out.Routes == nil {
		out.Routes = []route.Route{}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}

	return nil
}
