package registry

import (
	"reflect"
	"strings"
)

var registry Registry

// RegisterType adds a type to the global component registry. This should not need to be called directly, as the
// component scanning mechanism should generate calls to this function.
func RegisterType(t reflect.Type) {
	registry.RegisterType(t)
}

// Option to the ListTypes function.
type Option func(reflect.Type) bool

// ListTypes queries the global component registry. Without any options, returns all types. This will copy the registry,
// so using the options to filter the returned types is preferred.
func ListTypes(options ...Option) []reflect.Type {
	return registry.ListTypes(options...)
}

// InPackage provides the ability to filter registered types to those that are in a particular package or set of
// packages. If the package path ends with "/...", as in the standard go tools, this will search all packages rooted at
// the path before the "/...".
func InPackage(pkg string) Option {
	if pkg == "..." {
		return func(t reflect.Type) bool {
			return true
		}
	}
	if strings.HasSuffix(pkg, "/...") {
		return checkPrefix(pkg[:len(pkg)-len("/...")])
	}
	return func(t reflect.Type) bool {
		return pkg == t.PkgPath()
	}
}

func checkPrefix(pkg string) Option {
	return func(t reflect.Type) bool {
		p := t.PkgPath()
		if !strings.HasPrefix(p, pkg) {
			return false
		}
		// Check that this isn't matching another package the name of which happens to begin with the name of the
		// package we are interested in by checking that the relevant prefix is followed by '/'.
		// This is safe because if HasPrefix == true and len(pkg) != len(p) then there must be more characters in p.
		return len(pkg) == len(p) || p[len(pkg)] == '/'
	}
}

// AssignableTo provides the ability to filter registered types to those types that are assignable to the provided type,
// which will typically be a marker interface type, although any interface type will do.
func AssignableTo(intf reflect.Type) Option {
	return func(t reflect.Type) bool {
		return t.AssignableTo(intf)
	}
}
