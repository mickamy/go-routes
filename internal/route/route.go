// Package route defines the framework-agnostic representation of an extracted
// HTTP route and the helpers used to assemble request paths from nested groups.
package route

import "strings"

// VerbUnknown marks a route whose HTTP method could not be determined, such as
// a legacy net/http handler registered without a method pattern.
const VerbUnknown = "*"

// Route is a single URL the analyzed application exposes.
type Route struct {
	Framework string `json:"framework,omitempty"`
	Verb      string `json:"verb"`
	Path      string `json:"path"`
	Handler   string `json:"handler"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

// JoinPath concatenates a group prefix and a route segment into a single clean
// path, collapsing duplicated slashes so "/api/" + "/posts" yields "/api/posts".
// A trailing slash on the segment is preserved, since it is semantically
// significant in net/http (a subtree match) and most routers.
func JoinPath(prefix, segment string) string {
	if segment == "" {
		return prefix
	}

	joined := strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(segment, "/")
	if joined == "/" {
		return "/"
	}
	if strings.HasSuffix(segment, "/") {
		return joined
	}

	return strings.TrimRight(joined, "/")
}
