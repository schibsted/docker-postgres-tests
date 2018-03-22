package webapp

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
)

// AddFuncs adds the package's template functions.
func AddFuncs(t *template.Template, router *mux.Router) {
	t.Funcs(template.FuncMap{
		"path":  RoutePath(router),
		"url":   RouteURL(router),
		"cycle": Cycle,
	})
}

// Cycle cycles among the given strings given an index.
func Cycle(i int, arg0 string, args ...string) string {
	if j := i % (len(args) + 1); j != 0 {
		return args[j-1]
	}
	return arg0
}

// RoutePath returns a function suitable for a template that returns the path
// for a route and its parameters.
func RoutePath(router *mux.Router) func(string, ...interface{}) (template.URL, error) {
	return func(name string, pairs ...interface{}) (template.URL, error) {
		route := router.Get(name)
		if route == nil {
			return "", fmt.Errorf("path: no such route %q", name)
		}
		spairs := make([]string, len(pairs))
		for i := range pairs {
			spairs[i] = fmt.Sprint(pairs[i])
		}
		u, err := route.URLPath(spairs...)
		return template.URL(u.Path), err
	}
}

// RouteURL returns a function suitable for a template that returns the full URL
// for a route and its parameters.
func RouteURL(router *mux.Router) func(string, ...interface{}) (template.URL, error) {
	return func(name string, pairs ...interface{}) (template.URL, error) {
		route := router.Get(name)
		if route == nil {
			return "", fmt.Errorf("path: no such route %q", name)
		}
		spairs := make([]string, len(pairs))
		for i := range pairs {
			spairs[i] = fmt.Sprint(pairs[i])
		}
		u, err := route.URL(spairs...)
		return template.URL(u.String()), err
	}
}
