package registry

import "reflect"

// Registry is captured as a type here so that you can stub out implementations of it in your
// tests.
type Registry struct {
	types []reflect.Type
}

// RegisterType adds a type to the given registry.
func (r *Registry) RegisterType(t reflect.Type) {
	r.types = append(r.types, t)
}

// ListType lists the types that are stored in the registry.
func (r *Registry) ListTypes(options ...Option) []reflect.Type {
	var res []reflect.Type
	for _, t := range r.types {
		if testOptions(t, options) {
			res = append(res, t)
		}
	}
	return res
}

func testOptions(t reflect.Type, options []Option) bool {
	for _, option := range options {
		if !option(t) {
			return false
		}
	}
	return true
}

