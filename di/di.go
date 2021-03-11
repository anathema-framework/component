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

// Configuration is a builder for furnishing contexts.
type Configuration struct {
	factories []factory
}

// A Factory function is called to create a new value of a particular type.
type Factory func(ctx context.Context) (reflect.Value, error)

// AddFactory registers a factory function for a particular type. During furnishing, this factory will be used for types
// to which the given type is assignable.
func (c *Configuration) AddFactory(creating reflect.Type, impl Factory) {
	c.factories = append(c.factories, factory{creating: creating, impl: impl})
}

// Install the configuration into the provided context. This causes a copy of the configuration to be made, effectively
// freezing the configuration in the resulting context.
func (c *Configuration) Install(ctx context.Context) context.Context {
	factories := make([]factory, len(c.factories))
	copy(factories, c.factories)

	return context.WithValue(ctx, furnisherKey{}, &furnisher{
		factories: factories,
	})
}

// Furnish a reference to a value from the provided context. The
//
// The context must be configured, which is to say that it is a context that is in some way derived from the result of
// calling Configuration.Install. If this is not the case, ErrNoFurnisher is returned.
//
// The furnishing of a value is based primarily on the type of value being furnished. Any types that have a factory
// configured for them will use that factory. For those that do not, a set of generic rules are followed. These are
// given below. Note that when more than one factory matches, ErrTooManyFactories is returned, but you can avoid this
// using a slice (see below).
//
// Contexts (i.e. context.Context) are furnished with the current context (the one passed into Furnish).
//
// Pointers are dereferenced, allocating if they are nil, and the referent is furnished.
//
// Structs have all their exported fields furnished, excluding those with tags.
//
// Slices will be furnished using all the factories that produce types assignable to the slice's element type.
//
// Functions will be called with all the arguments furnished from the current context.
func Furnish(ctx context.Context, ref interface{}) error {
	return FurnishValue(ctx, reflect.ValueOf(ref))
}

// FurnishValue will furnish the provided value from the provided context. See Furnish for more information about the
// furnishing process.
func FurnishValue(ctx context.Context, ref reflect.Value) error {
	f := getFurnisher(ctx)
	if f == nil {
		return ErrNoFurnisher
	}

	return f.furnish(ctx, ref)
}

// FurnishArgs will create a slice containing all the arguments for the provided function, furnished from the provided
// context. See Furnish for a description of the furnishing process.
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
