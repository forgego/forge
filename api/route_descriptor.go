package api

import (
	"net/http"
	"sort"
	"strings"
)

// RouteDescriptor describes a route mounted by an API Router.
type RouteDescriptor struct {
	Method   string
	Path     string
	Resource string
	Action   string
	Detail   bool
}

// Routes returns a slice of all route descriptors mounted by the router.
func (r *Router) Routes() []RouteDescriptor {
	var descriptors []RouteDescriptor

	resources := make([]string, 0, len(r.routes))
	for res := range r.routes {
		resources = append(resources, res)
	}
	sort.Strings(resources)

	for _, res := range resources {
		descriptors = append(descriptors, r.resourceDescriptors(res)...)
	}

	descriptors = append(descriptors, r.actionDescriptors()...)

	for _, cr := range r.customRoutes {
		cleanPath := "/" + strings.TrimPrefix(cr.path, "/")
		fullPath := r.prefix + cleanPath
		descriptors = append(descriptors, RouteDescriptor{
			Method:   cr.method,
			Path:     fullPath,
			Resource: "",
			Action:   "",
			Detail:   false,
		})
	}

	return descriptors
}

func (r *Router) resourceDescriptors(resource string) []RouteDescriptor {
	cleanPrefix := strings.TrimSuffix(r.prefix, "/")
	cleanResource := strings.Trim(resource, "/")
	var basePath string
	if cleanPrefix == "" {
		basePath = "/" + cleanResource
	} else {
		basePath = cleanPrefix + "/" + cleanResource
	}
	collectionPath := basePath + "/"
	detailPath := basePath + "/{id}"

	return []RouteDescriptor{
		{
			Method:   http.MethodGet,
			Path:     collectionPath,
			Resource: resource,
			Action:   "list",
			Detail:   false,
		},
		{
			Method:   http.MethodPost,
			Path:     collectionPath,
			Resource: resource,
			Action:   "create",
			Detail:   false,
		},
		{
			Method:   http.MethodGet,
			Path:     detailPath,
			Resource: resource,
			Action:   "retrieve",
			Detail:   true,
		},
		{
			Method:   http.MethodPut,
			Path:     detailPath,
			Resource: resource,
			Action:   "update",
			Detail:   true,
		},
		{
			Method:   http.MethodPatch,
			Path:     detailPath,
			Resource: resource,
			Action:   "partial_update",
			Detail:   true,
		},
		{
			Method:   http.MethodDelete,
			Path:     detailPath,
			Resource: resource,
			Action:   "destroy",
			Detail:   true,
		},
		{
			Method:   http.MethodOptions,
			Path:     collectionPath,
			Resource: resource,
			Action:   "metadata",
			Detail:   false,
		},
	}
}

func (r *Router) actionDescriptors() []RouteDescriptor {
	cleanPrefix := strings.TrimSuffix(r.prefix, "/")
	var descriptors []RouteDescriptor
	for _, act := range r.actions {
		cleanResource := strings.Trim(act.resource, "/")
		cleanActionPath := strings.Trim(act.cfg.URLPath, "/")
		var fullPath string
		if act.cfg.Detail {
			fullPath = cleanPrefix + "/" + cleanResource + "/{id}/" + cleanActionPath
		} else {
			fullPath = cleanPrefix + "/" + cleanResource + "/" + cleanActionPath
		}
		for _, method := range act.cfg.Methods {
			descriptors = append(descriptors, RouteDescriptor{
				Method:   method,
				Path:     fullPath,
				Resource: act.resource,
				Action:   act.name,
				Detail:   act.cfg.Detail,
			})
		}
	}
	return descriptors
}
