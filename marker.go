package component

import (
	"go/types"
	"reflect"
)

// Marker interface that represents something to be scanned for.
//
// To have anathema detect your types, embed them with this marker type. You can
// embed them in other interface types to turn them into component markers too.
// This is how, for example, the Resource marker works.
type Marker interface {
	marker()
}

var markerType = reflect.TypeOf(new(Marker)).Elem()

var markerPkg = types.NewPackage(markerType.PkgPath(), "")

var markerCompileType = types.NewNamed(
	types.NewTypeName(0, markerPkg, markerType.Name(), nil),
	types.NewInterfaceType([]*types.Func{types.NewFunc(
		0, markerPkg, "marker",
		types.NewSignature(nil, nil, nil, false),
	)}, nil).Complete(),
	nil,
)

// Type returns the reflect type of the component marker.
func Type() reflect.Type {
	return markerType
}

// CompileType returns the compile-time type of the component marker.
func CompileType() types.Type {
	return markerCompileType
}

// Tag retrieves the tag for the given component.
func Tag(m Marker) reflect.StructTag {
	return TypeTag(reflect.TypeOf(m))
}

// TypeTag retrieves the tag for a reflect.Type representation of a component.
func TypeTag(t reflect.Type) reflect.StructTag {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var res reflect.StructTag
	if t.Kind() != reflect.Struct {
		return res
	}
	if !t.AssignableTo(markerType) {
		return res
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Type.AssignableTo(markerType) {
			continue
		}
		if res != "" {
			res += " "
		}
		res += f.Tag
	}
	return res

}
