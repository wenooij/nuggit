package v1alpha

import (
	"context"
	"fmt"
	"regexp"
)

func (x *Regex) Bind(edges []Edge) error {
	for _, e := range edges {
		switch res := e.Result.(type) {
		case string:
			x.Pattern = res
		default:
			return fmt.Errorf("unexpected type for Regex: %T", res)
		}
	}
	if x.Pattern == "" {
		return fmt.Errorf("empty pattern in Regex")
	}
	return nil
}

func (x *Regex) Run(ctx context.Context) (any, error) {
	r, err := regexp.Compile(x.Pattern)
	if err != nil {
		return nil, err
	}
	return r, nil
}
