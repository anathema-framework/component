package load

import (
	"context"
	"github.com/anathema-framework/assert"
	"github.com/anathema-framework/component/a"
	"github.com/anathema-framework/component/di"
	"github.com/anathema-framework/component/registry"
	"github.com/anathema-framework/component/scope"
	"reflect"
	"testing"
)

type exampleService struct {
	a.Service

	Value int
}

func (s *exampleService) GetValue() int {
	return s.Value
}

type exampleProvider struct {
	a.Provider
}

func (p *exampleProvider) GetInt() int {
	return 13
}

func TestLoad(t *testing.T) {
	c := new(di.Configuration)

	r := new(registry.Registry)
	r.RegisterType(reflect.TypeOf(new(exampleService)).Elem())
	r.RegisterType(reflect.TypeOf(new(exampleProvider)).Elem())

	l := loader{c, r.ListTypes, nil}
	l.services()
	l.providers()
	assert.NoError(t, l.err)

	ctx := c.Install(scope.Enter(context.Background(), ""))

	var service interface {
		GetValue() int
	}
	assert.NoError(t, di.Furnish(ctx, &service))
	assert.Equal(t, service.GetValue(), 13)

}
