package di

import (
	"context"
	"fmt"
	"github.com/anathema-framework/component"
	"reflect"
)

type factory struct {
	creating reflect.Type
	impl     Factory
}

type furnisher struct {
	factories []factory
}

type furnisherKey struct{}

func getFurnisher(ctx context.Context) *furnisher {
	f := ctx.Value(furnisherKey{})
	if f == nil {
		return nil
	}
	return f.(*furnisher)
}

var contextType = reflect.TypeOf(new(context.Context)).Elem()

func (f *furnisher) furnish(ctx context.Context, target reflect.Value) error {
	if target.Type() == contextType {
		target.Set(reflect.ValueOf(ctx))
	}

	found := f.findFactories(target.Type())

	var err error
	switch len(found) {
	case 0:
		err = f.furnishFallback(ctx, target)
	case 1:
		err = f.furnishFromFactory(ctx, target, found[0])
	default:
		err = ErrTooManyFactories
	}

	return f.postConstruct(ctx, target, err)
}

func (f *furnisher) findFactories(t reflect.Type) []factory {
	var res []factory
	for _, candidate := range f.factories {
		if candidate.creating.AssignableTo(t) {
			res = append(res, candidate)
		}
	}
	return res
}

func (f *furnisher) furnishFromFactory(ctx context.Context, target reflect.Value, from factory) error {
	v, err := from.impl(ctx)
	if err != nil {
		return err
	}
	target.Set(v)
	return nil
}

func (f *furnisher) furnishFallback(ctx context.Context, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Ptr:
		return f.furnishPtr(ctx, target)

	case reflect.Struct:
		return f.furnishStruct(ctx, target)

	case reflect.Slice:
		return f.furnishSlice(ctx, target)

	case reflect.Func:
		return f.furnishCall(ctx, target)
	}

	return ErrUnsupportedType
}

func (f *furnisher) furnishPtr(ctx context.Context, target reflect.Value) error {
	if target.IsNil() {
		target.Set(reflect.New(target.Type().Elem()))
	}
	return f.furnish(ctx, target.Elem())
}

func (f *furnisher) furnishStruct(ctx context.Context, target reflect.Value) error {
	t := target.Type()

	for i := t.NumField() - 1; i >= 0; i-- {
		fp := t.Field(i)
		if fp.PkgPath != "" {
			continue
		}
		if fp.Type.AssignableTo(component.Type()) {
			continue
		}

		fv := target.FieldByIndex(fp.Index)
		err := f.furnish(ctx, fv)
		if err != nil {
			return fmt.Errorf("field %s: %w", fp.Name, err)
		}
	}

	return nil
}

func (f *furnisher) furnishSlice(ctx context.Context, target reflect.Value) error {
	found := f.findFactories(target.Type().Elem())
	res := reflect.MakeSlice(target.Type(), len(found), cap(found))
	for i, fc := range found {
		err := f.furnishFromFactory(ctx, res.Index(i), fc)
		if err != nil {
			return err
		}
	}
	target.Set(res)
	return nil
}

func (f *furnisher) furnishCall(ctx context.Context, target reflect.Value) error {
	in, err := FurnishArgs(ctx, target)
	if err != nil {
		return err
	}
	out := target.Call(in)
	if len(out) == 0 || out[0].IsNil() {
		return nil
	}
	if e, ok := out[0].Interface().(error); ok {
		return e
	}
	return nil
}

func (f *furnisher) postConstruct(ctx context.Context, target reflect.Value, err error) error {
	if err == nil {
		if m := target.MethodByName("Inject"); m.IsValid() {
			err = f.furnishCall(ctx, m)
		}
	}
	if err != nil {
		return fmt.Errorf("furnishing %v: %w", target.Type(), err)
	}
	return nil
}
