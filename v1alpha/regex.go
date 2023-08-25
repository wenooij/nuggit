package v1alpha

import (
	"context"
	"fmt"
	"regexp"

	"github.com/wenooij/nuggit/runtime"
)

func (x *Regex) Bind(e runtime.Edge) error {
	switch res := e.Result.(type) {
	case string:
		x.Pattern = res
	default:
		return fmt.Errorf("unexpected type for Regex: %T", res)
	}
	return nil
}

func (x *Regex) Run(ctx context.Context) (any, error) {
	if x.Pattern == "" {
		return nil, fmt.Errorf("empty pattern in Regex")
	}
	r, err := regexp.Compile(x.Pattern)
	if err != nil {
		return nil, err
	}
	return r, nil
}
