package v1alpha

import (
	"context"
)

func (x *HTTP) Bind([]Edge) error {
	return nil
}

func (r *HTTP) Run(ctx context.Context) (any, error) {
	panic("HTTP.Run not implemented")
}
