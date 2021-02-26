package di

import (
	"context"
	"errors"
	"reflect"
)

var (
	ErrNoFurnisher      = errors.New("no furnisher in the current context")
	ErrUnsupportedType  = errors.New("type cannot be furnished")
	ErrTooManyFactories = errors.New("multiple factories located")
)

// Configuration is a builder for furnishers.
type Configuration struct {
	factories []factory
}

type Factory func(ctx context.Context) (reflect.Value, error)

func Furnish(ctx context.Context, ref interface{}) error {
	return FurnishValue(ctx, reflect.ValueOf(ref))
}

func FurnishValue(ctx context.Context, ref reflect.Value) error {
	f := getFurnisher(ctx)
	if f == nil {
		return ErrNoFurnisher
	}

	return f.furnish(ctx, ref)
}

func FurnishArgs(ctx context.Context, fn reflect.Value) ([]reflect.Value, error) {
	f := getFurnisher(ctx)
	if f == nil {
		return nil, ErrNoFurnisher
	}
	t := fn.Type()
	in := make([]reflect.Value, t.NumIn())

	for i := range in {
		in[i] = reflect.New(t.In(i)).Elem()
		err := f.furnish(ctx, in[i])
		if err != nil {
			return nil, err
		}
	}

	return in, nil
}

func (c *Configuration) AddFactory(creating reflect.Type, impl Factory) {
	c.factories = append(c.factories, factory{creating: creating, impl: impl})
}

func (c *Configuration) Install(ctx context.Context) context.Context {
	factories := make([]factory, len(c.factories))
	copy(factories, c.factories)

	return context.WithValue(ctx, furnisherKey{}, &furnisher{
		factories: factories,
	})
}
