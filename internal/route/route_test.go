package route_test

import (
	"testing"

	"github.com/mickamy/go-routes/internal/route"
)

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prefix  string
		segment string
		want    string
	}{
		{name: "empty prefix", prefix: "", segment: "/posts", want: "/posts"},
		{name: "root prefix", prefix: "/", segment: "/posts", want: "/posts"},
		{name: "nested prefix", prefix: "/api", segment: "/posts", want: "/api/posts"},
		{name: "trailing slash prefix", prefix: "/api/", segment: "/posts", want: "/api/posts"},
		{name: "segment without leading slash", prefix: "/api", segment: "posts", want: "/api/posts"},
		{name: "root only", prefix: "", segment: "/", want: "/"},
		{name: "prefix only", prefix: "/api", segment: "", want: "/api"},
		{name: "param segment", prefix: "/posts", segment: "/:id", want: "/posts/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := route.JoinPath(tt.prefix, tt.segment); got != tt.want {
				t.Errorf("JoinPath(%q, %q) = %q, want %q", tt.prefix, tt.segment, got, tt.want)
			}
		})
	}
}
