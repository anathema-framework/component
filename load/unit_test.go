package load

import (
	"github.com/anathema-framework/assert"
	"reflect"
	"testing"
)

func TestProviderCheck(t *testing.T) {
	for _, test := range []struct {
		name string
		method interface{}
		err error
	} {
		{
			"no return",
			func() {},
			ErrBadProvider,
		},
		{
			"only error return",
			func() error { return nil},
			ErrBadProvider,
		},
		{
			"two error returns",
			func() (error, error) { return nil, nil},
			ErrBadProvider,
		},
		{
			"too many returns",
			func() (string, string, string) { return "", "", ""},
			ErrBadProvider,
		},
		{
			"non-error second return",
			func() (string, string) { return "", ""},
			ErrBadProvider,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			mt := reflect.TypeOf(test.method)
			err := checkProvider(mt)
			assert.ErrorIs(t, err, test.err)
		})
	}
}
