package v1alpha

import (
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/runtime"
)

type Factory struct{}

func (*Factory) NewRunner(n nuggit.Node) (runtime.Runner, error) {
	var op runtime.Runner
	switch n.Op {
	case "Chromedp":
		op = &Chromedp{}
	case "Const":
		op = &Const{}
	case "HTTP":
		op = &HTTP{}
	case "Sink":
		op = &Sink{}
	case "Source":
		op = &Source{}
	case "String":
		op = &String{}
	case "Var":
		op = &Var{}
	default:
		return nil, fmt.Errorf("NewRunner: Runner is not defined for Op: %q: %q", n.Key, n.Op)
	}
	if len(n.Data) > 0 {
		if err := json.Unmarshal(n.Data, op); err != nil {
			return nil, err
		}
	}
	return op, nil
}
