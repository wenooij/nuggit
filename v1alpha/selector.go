package v1alpha

import (
	"context"

	"github.com/andybalholm/cascadia"
)

func (x *Selector) Bind([]Edge) error {
	// TODO(wes): Implement selector from input string.
	return nil
}

func (x *Selector) Run(ctx context.Context) (any, error) {
	sel, err := cascadia.Compile(x.Selector)
	if err != nil {
		return nil, err
	}
	if x.All {
		return cascadia.QueryAll(x.Node, sel), nil
	}
	return cascadia.Query(x.Node, sel), nil
}
