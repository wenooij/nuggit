package v1alpha

import (
	"context"
	"fmt"
	"regexp"

	"github.com/wenooij/nuggit/runtime"
)

func (x *Find) Bind(e runtime.Edge) error {
	// TODO(wes): Bind edges for Find.
	return nil
}

func (x *Find) Run(ctx context.Context) (any, error) {
	if x.Regex == nil {
		return nil, fmt.Errorf("Find must have a Regex")
	}
	if x.Sink == nil {
		return nil, fmt.Errorf("Find must have a Sink")
	}
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
