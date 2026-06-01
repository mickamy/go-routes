// Package loader wraps golang.org/x/tools/go/packages to load a Go module with
// full type information, which the framework visitors rely on to resolve
// handler symbols and group receiver types.
package loader

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Loaded is the result of loading a target directory: the typed packages, the
// shared file set, and the root used to relativize file paths in output.
type Loaded struct {
	Packages []*packages.Package
	Fset     *token.FileSet
	Root     string
}

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedModule

// Load type-checks every package under dir (recursively) and returns them. Type
// errors in the target do not abort the load; the syntax and partial type info
// are still returned so analysis can proceed on a best-effort basis.
func Load(dir string) (Loaded, error) {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode:  loadMode,
		Dir:   dir,
		Fset:  fset,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return Loaded{}, fmt.Errorf("loading packages from %q: %w", dir, err)
	}
	if len(pkgs) == 0 {
		return Loaded{}, fmt.Errorf("no Go packages found under %q", dir)
	}

	return Loaded{
		Packages: pkgs,
		Fset:     fset,
		Root:     moduleRoot(pkgs, dir),
	}, nil
}

// RelPath turns an absolute file path into one relative to the module root,
// falling back to the original path when it lies outside the root.
func (l Loaded) RelPath(abs string) string {
	if l.Root == "" {
		return abs
	}

	rel, err := filepath.Rel(l.Root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return abs
	}

	return rel
}

func moduleRoot(pkgs []*packages.Package, dir string) string {
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Dir != "" {
			return pkg.Module.Dir
		}
	}

	if abs, err := filepath.Abs(dir); err == nil {
		return abs
	}

	return dir
}
