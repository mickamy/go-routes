package framework_test

import (
	"testing"

	"github.com/mickamy/go-routes/internal/framework"
	"github.com/mickamy/go-routes/internal/loader"
	"github.com/mickamy/go-routes/internal/route"
)

func TestNetHTTP(t *testing.T) {
	t.Parallel()

	l, err := loader.Load("testdata/nethttp")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	routes, err := framework.Extract(l, "net/http")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	type key struct {
		verb, path, handler, file string
	}
	got := map[key]bool{}
	for _, r := range routes {
		if r.Line <= 0 {
			t.Errorf("route %s %s has non-positive line %d", r.Verb, r.Path, r.Line)
		}
		got[key{r.Verb, r.Path, r.Handler, r.File}] = true
	}

	want := []key{
		{"GET", "/", "handlers.Index", "handlers/handlers.go"},
		{"GET", "/posts", "handlers.List", "handlers/handlers.go"},
		{"POST", "/posts", "handlers.Create", "handlers/handlers.go"},
		{route.VerbUnknown, "/legacy", "handlers.Index", "handlers/handlers.go"},
		{"DELETE", "/posts/{id}", "func", "main.go"},
		{"GET", "/healthz", "handlers.Index", "handlers/handlers.go"},
		{"GET", "/api/comments", "handlers.List", "handlers/handlers.go"},
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing route %+v\ngot: %+v", w, routes)
		}
	}

	if len(routes) != len(want) {
		t.Errorf("got %d routes, want %d: %+v", len(routes), len(want), routes)
	}
}
