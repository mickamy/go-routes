package render_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mickamy/go-routes/internal/render"
	"github.com/mickamy/go-routes/internal/route"
)

func sampleRoutes() []route.Route {
	return []route.Route{
		{Framework: "echo", Verb: "GET", Path: "/", Handler: "home.Index", File: "handlers/home.go", Line: 12},
		{Framework: "echo", Verb: "POST", Path: "/posts", Handler: "posts.Create", File: "handlers/posts.go", Line: 48},
	}
}

func TestRenderTable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := render.Render(&buf, sampleRoutes(), render.Table); err != nil {
		t.Fatalf("render: %v", err)
	}

	out := buf.String()
	wants := []string{
		"VERB", "GET", "/posts", "posts.Create",
		"handlers/posts.go:48", "2 routes in 2 files · echo",
	}
	for _, want := range wants {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTableEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := render.Render(&buf, nil, render.Table); err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "no routes found" {
		t.Errorf("got %q, want %q", got, "no routes found")
	}
}

func TestRenderJSONSingleFramework(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := render.Render(&buf, sampleRoutes(), render.JSON); err != nil {
		t.Fatalf("render: %v", err)
	}

	var out struct {
		Framework string `json:"framework"`
		Routes    []struct {
			Framework string `json:"framework"`
			Verb      string `json:"verb"`
			Path      string `json:"path"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}

	if out.Framework != "echo" {
		t.Errorf("top-level framework = %q, want echo", out.Framework)
	}
	if len(out.Routes) != 2 {
		t.Fatalf("got %d routes, want 2", len(out.Routes))
	}
	if out.Routes[0].Framework != "" {
		t.Errorf("per-route framework should be omitted for single-framework output, got %q", out.Routes[0].Framework)
	}
}

func TestRenderJSONMultiFramework(t *testing.T) {
	t.Parallel()

	routes := append(sampleRoutes(),
		route.Route{Framework: "chi", Verb: "GET", Path: "/api", Handler: "api.List", File: "api.go", Line: 3},
	)

	var buf bytes.Buffer
	if err := render.Render(&buf, routes, render.JSON); err != nil {
		t.Fatalf("render: %v", err)
	}

	var out struct {
		Framework string `json:"framework"`
		Routes    []struct {
			Framework string `json:"framework"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Framework != "" {
		t.Errorf("top-level framework should be empty for multi-framework output, got %q", out.Framework)
	}
	if out.Routes[0].Framework == "" {
		t.Errorf("per-route framework should be present for multi-framework output")
	}
}

func TestParseFormat(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"table", "json"} {
		if _, err := render.ParseFormat(s); err != nil {
			t.Errorf("ParseFormat(%q) errored: %v", s, err)
		}
	}
	if _, err := render.ParseFormat("yaml"); err == nil {
		t.Error("ParseFormat(\"yaml\") should error")
	}
}
