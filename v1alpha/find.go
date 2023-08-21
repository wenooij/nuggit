package v1alpha

import (
	"context"
	"fmt"
	"regexp"

	"github.com/wenooij/nuggit/runtime"
)

func (x *Find) Bind([]runtime.Edge) error {
	if x.Regex == nil {
		return fmt.Errorf("Find must have a Regex")
	}
	if x.Sink == nil {
		return fmt.Errorf("Find must have a Sink")
	}
	// TODO(wes): Validate other types of Find ops here.
	return nil
}

func (x *Find) Run(ctx context.Context) (any, error) {
	// TODO(wes): Support reverse.
	rgx, err := regexp.Compile(x.Regex.Pattern)
	if err != nil {
		return nil, err
	}
	var matches [][]int
	switch {
	case x.All && x.Submatch:
		matches = rgx.FindAllSubmatchIndex(x.Sink.Bytes, -1)
	case x.All:
		matches = rgx.FindAllIndex(x.Sink.Bytes, -1)
	case x.Submatch:
		matches = [][]int{rgx.FindSubmatchIndex(x.Sink.Bytes)}
	default:
		matches = [][]int{rgx.FindIndex(x.Sink.Bytes)}
	}

	return matches, nil
}
