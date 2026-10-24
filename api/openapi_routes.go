package api

import "github.com/forgego/forge/api/docs"

// DocsRoutes converts a slice of RouteDescriptors into docs.Routes.
// It copies Method, Path, Resource, Action, and Detail. Fields and Security
// stay empty for callers to fill.
func DocsRoutes(descs []RouteDescriptor) []docs.Route {
	if descs == nil {
		return nil
	}
	routes := make([]docs.Route, len(descs))
	for i, d := range descs {
		routes[i] = docs.Route{
			Method:   d.Method,
			Path:     d.Path,
			Resource: d.Resource,
			Action:   d.Action,
			Detail:   d.Detail,
		}
	}
	return routes
}
