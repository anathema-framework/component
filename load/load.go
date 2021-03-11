package load

import (
	"context"
	"errors"
	"fmt"
	"github.com/anathema-framework/component"
	"github.com/anathema-framework/component/a"
	"github.com/anathema-framework/component/di"
	"github.com/anathema-framework/component/registry"
	"github.com/anathema-framework/component/scope"
	"reflect"
)

var (
	ErrBadProvider = errors.New("bad provider")
	ErrNoScope     = errors.New("missing scope")
)

// Services will load all the services that have been registered with the application into the provided Configuration.
func Services(c *di.Configuration) error {
	load := loader{c, registry.ListTypes, nil}
	load.services()
	load.providers()
	return load.err
}

func AddServiceFactory(c *di.Configuration, service reflect.Type) {
	c.AddFactory(reflect.PtrTo(service), func(ctx context.Context) (reflect.Value, error) {
		scopeName := component.TypeTag(service).Get("scope")
		s := scope.Retrieve(ctx, scopeName)
		if s == nil {
			return reflect.Value{}, fmt.Errorf("%s: %w", scopeName, ErrNoScope)
		}

		value, ok := s.Get(reflect.PtrTo(service))
		if !ok {
			value = reflect.New(service)
			s.Insert(reflect.PtrTo(service), value)

			err := di.FurnishValue(ctx, value.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
		}

		return value, nil
	})
}

func AddProviderFactories(c *di.Configuration, provider reflect.Type) error {
	p := provider
	if p.Kind() == reflect.Ptr {
		p = p.Elem()
	}
	AddServiceFactory(c, p)

	for i := 0; i < provider.NumMethod(); i++ {
		m := provider.Method(i)
		if m.PkgPath != "" {
			continue
		}
		if m.Name == "Inject" {
			continue
		}

		err := addProviderFactory(c, m)
		if err != nil {
			return err
		}
	}

	return nil
}

var (
	serviceType  = reflect.TypeOf(new(a.Service)).Elem()
	providerType = reflect.TypeOf(new(a.Provider)).Elem()
)

type loader struct {
	c   *di.Configuration
	src func(...registry.Option) []reflect.Type
	err error
}

func (l *loader) services() {
	if l.err != nil {
		return
	}

	for _, service := range l.src(registry.AssignableTo(serviceType)) {
		if service.Kind() != reflect.Struct {
			continue
		}
		l.addServiceFactory(service)
	}
}

func (l *loader) providers() {
	if l.err != nil {
		return
	}

	for _, provider := range l.src(registry.AssignableTo(providerType)) {
		if provider.Kind() != reflect.Struct {
			continue
		}

		provider = reflect.PtrTo(provider)
		l.err = AddProviderFactories(l.c, provider)
		if l.err != nil {
			return
		}
	}
}

func (l *loader) addServiceFactory(service reflect.Type) {
	AddServiceFactory(l.c, service)
}

var errType = reflect.TypeOf(new(error)).Elem()

func addProviderFactory(c *di.Configuration, m reflect.Method) error {
	err := checkProvider(m.Type)
	if err != nil {
		return fmt.Errorf("%s: %w", m.Name, err)
	}

	c.AddFactory(m.Type.Out(0), func(ctx context.Context) (reflect.Value, error) {
		in, err := di.FurnishArgs(ctx, m.Func)
		if err != nil {
			return reflect.Value{}, err
		}

		out := m.Func.Call(in)

		if len(out) == 1 || out[1].IsNil() {
			return out[0], nil
		}

		return reflect.Value{}, out[1].Interface().(error)
	})

	return nil
}

func checkProvider(mt reflect.Type) error {
	switch mt.NumOut() {
	case 1:
		if mt.Out(0).AssignableTo(errType) {
			return ErrBadProvider
		}
	case 2:
		if mt.Out(0).AssignableTo(errType) {
			return ErrBadProvider
		}
		if !mt.Out(1).AssignableTo(errType) {
			return ErrBadProvider
		}
	default:
		return ErrBadProvider
	}
	return nil
}

