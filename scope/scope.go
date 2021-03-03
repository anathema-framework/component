package scope

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"io"
	"reflect"
	"sync"
)

type Scope struct {
	prev   *Scope
	name   string
	values map[reflect.Type]reflect.Value
	lock   sync.RWMutex
}

type scopeKey struct{}

func Enter(ctx context.Context, name string) context.Context {
	var prev *Scope
	if p := ctx.Value(scopeKey{}); p != nil {
		prev = p.(*Scope)
	}
	scope := &Scope{
		prev:   prev,
		name:   name,
		values: map[reflect.Type]reflect.Value{},
	}
	return context.WithValue(ctx, scopeKey{}, scope)
}

func Retrieve(ctx context.Context, name string) *Scope {
	v := ctx.Value(scopeKey{})
	if v == nil {
		return nil
	}
	return v.(*Scope).find(name)
}

func (s *Scope) find(name string) *Scope {
	if s == nil {
		return nil
	}
	if s.name == name {
		return s
	}
	return s.prev.find(name)
}

func (s *Scope) Get(t reflect.Type) (reflect.Value, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	v, ok := s.values[t]
	return v, ok
}

func (s *Scope) Insert(t reflect.Type, v reflect.Value) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.values[t] = v
}

func (s *Scope) Ensure(t reflect.Type, fv func() reflect.Value) reflect.Value {
	v, ok := s.Get(t)
	if !ok {
		v = fv()
		s.Insert(t, v)
	}
	return v
}

func (s *Scope) Close() error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var errs error

	for _, v := range s.values {
		if v, ok := v.Interface().(io.Closer); ok {
			err := v.Close()
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs
}
