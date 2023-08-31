package v1alpha

import (
	"context"
	"fmt"
	"regexp"
)

// Regex defines a Go-style regular expression.
//
// Pattern should be a string input the regular expression.
//
// The pattern can incorporate steps and variables using
// step inputs.
//
// Syntax: https://golang.org/s/re2syntax.
type Regex struct {
	Pattern string `json:"pattern,omitempty"`
}

func (x *Regex) Compile() (*regexp.Regexp, error) {
	r, err := regexp.Compile(x.Pattern)
	if err != nil {
		return nil, err
	}
	return r, err
}

func (x *Regex) Validate() error {
	if x.Pattern == "" {
		return fmt.Errorf("missing Pattern")
	}
	return nil
}

func (x *Regex) Run(ctx context.Context) (any, error) {
	return x, nil
}
