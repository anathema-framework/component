package di

import (
	"context"
	"github.com/anathema-framework/assert"
	"reflect"
	"testing"
)

type testTarget struct {
	x int
}

func (t *testTarget) Inject(x int) {
	t.x = x
}

func TestFurnish(t *testing.T) {
	c := new(Configuration)
	c.AddFactory(reflect.TypeOf(0), func(ctx context.Context) (reflect.Value, error) {
		return reflect.ValueOf(12), nil
	})

	ctx := c.Install(context.Background())

	var inject struct {
		I int
	}
	assert.NoError(t, Furnish(ctx, &inject))
	assert.Equal(t, inject.I, 12)

	var inject2 testTarget
	assert.NoError(t, Furnish(ctx, &inject2))
	assert.Equal(t, inject2.x, 12)

}
