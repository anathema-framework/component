package registry

import (
	"github.com/anathema-framework/assert"
	"reflect"
	"testing"
)

func TestOptions(t *testing.T) {
	for _, test := range []struct {
		name   string
		option Option
		typ    reflect.Type
		match  bool
	}{
		{
			name:   "type in package",
			option: InPackage("github.com/anathema-framework/component/registry"),
			typ:    reflect.TypeOf(new(Registry)).Elem(),
			match:  true,
		},
		{
			name: "type not in package",
			option: InPackage("github.com/anathema-framework/component/registry"),
			typ:    reflect.TypeOf(new(string)).Elem(),
			match:  false,
		},
		{
			name:   "type in package range",
			option: InPackage("github.com/anathema-framework/..."),
			typ:    reflect.TypeOf(new(Registry)).Elem(),
			match:  true,
		},
		{
			name: "type not in package range",
			option: InPackage("github.com/anathema-framework/..."),
			typ:    reflect.TypeOf(new(string)).Elem(),
			match:  false,
		},
		{
			name: "type in any package",
			option: InPackage("..."),
			typ:    reflect.TypeOf(new(string)).Elem(),
			match:  true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.match, test.option(test.typ))
		})
	}
}
