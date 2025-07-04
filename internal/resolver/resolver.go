package resolver

import (
	"fmt"
	"sort"

	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/javanhut/bifrost/internal/version"
)

type Package struct {
	Name         string
	Version      *version.Version
	Dependencies map[string]version.Constraint
}

type Resolution struct {
	Packages map[string]*Package
}

type Resolver struct {
	packages map[string][]*Package // name -> available versions
}

func New() *Resolver {
	return &Resolver{
		packages: make(map[string][]*Package),
	}
}

func (r *Resolver) AddPackage(pkg *Package) {
	r.packages[pkg.Name] = append(r.packages[pkg.Name], pkg)
}

func (r *Resolver) Resolve(root *manifest.Manifest) (*Resolution, error) {
	// Convert manifest dependencies to packages
	rootPkg := &Package{
		Name:         root.Package.Name,
		Version:      &version.Version{Major: 0, Minor: 0, Patch: 0},
		Dependencies: make(map[string]version.Constraint),
	}

	for name, constraintStr := range root.Dependencies {
		constraint, err := version.ParseConstraint(constraintStr)
		if err != nil {
			return nil, fmt.Errorf("invalid constraint for %s: %w", name, err)
		}
		rootPkg.Dependencies[name] = constraint
	}

	// Run the resolution algorithm
	resolved := make(map[string]*Package)
	if err := r.resolvePackage(rootPkg, resolved, nil); err != nil {
		return nil, err
	}

	// Remove the root package from results
	delete(resolved, rootPkg.Name)

	return &Resolution{Packages: resolved}, nil
}

func (r *Resolver) resolvePackage(pkg *Package, resolved map[string]*Package, stack []string) error {
	// Check for circular dependencies
	for _, name := range stack {
		if name == pkg.Name {
			return fmt.Errorf("circular dependency detected: %s", append(stack, pkg.Name))
		}
	}

	stack = append(stack, pkg.Name)

	for depName, constraint := range pkg.Dependencies {
		// Check if already resolved
		if existing, ok := resolved[depName]; ok {
			if !constraint.Satisfies(existing.Version) {
				return fmt.Errorf("version conflict for %s: %s requires %s, but %s is already resolved",
					depName, pkg.Name, constraint, existing.Version)
			}
			continue
		}

		// Find compatible version
		candidates := r.packages[depName]
		if len(candidates) == 0 {
			return fmt.Errorf("package not found: %s", depName)
		}

		// Sort by version (newest first)
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Version.Compare(candidates[j].Version) > 0
		})

		// Find first compatible version
		var selected *Package
		for _, candidate := range candidates {
			if constraint.Satisfies(candidate.Version) {
				selected = candidate
				break
			}
		}

		if selected == nil {
			return fmt.Errorf("no compatible version found for %s with constraint %s", depName, constraint)
		}

		// Add to resolved
		resolved[depName] = selected

		// Recursively resolve dependencies
		if err := r.resolvePackage(selected, resolved, stack); err != nil {
			return err
		}
	}

	return nil
}

// GetResolutionOrder returns packages in the order they should be installed
func (res *Resolution) GetResolutionOrder() []*Package {
	// Build dependency graph
	graph := make(map[string][]string)
	packages := make(map[string]*Package)

	for name, pkg := range res.Packages {
		packages[name] = pkg
		graph[name] = make([]string, 0, len(pkg.Dependencies))
		for depName := range pkg.Dependencies {
			if _, ok := res.Packages[depName]; ok {
				graph[name] = append(graph[name], depName)
			}
		}
	}

	// Topological sort
	var order []*Package
	visited := make(map[string]bool)
	var visit func(string)

	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		for _, dep := range graph[name] {
			visit(dep)
		}

		if pkg, ok := packages[name]; ok {
			order = append(order, pkg)
		}
	}

	for name := range packages {
		visit(name)
	}

	return order
}
