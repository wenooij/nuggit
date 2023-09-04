package v1alpha

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

type Factory struct{}

func (Factory) New(n nuggit.Node) (any, error) {
	opFn, ok := factoryMapping[n.Op]
	if !ok {
		return nil, fmt.Errorf("v1alpha.Factory.NewRunner: Node %q: Runner is not defined for Op: %q", n.Key, n.Op)
	}
	return opFn(), nil
}

var factoryMapping = map[string]func() any{
	"array":    func() any { return &[]any{} },
	"any":      func() any { return new(any) },
	"bool":     func() any { var t bool; return &t },
	"Chromedp": func() any { return &Chromedp{} },
	"Const":    func() any { return &Const{} },
	"HTML":     func() any { return &HTML{} },
	"HTTP":     func() any { return &HTTP{} },
	"map":      func() any { return &map[string]any{} },
	"null":     func() any { return nil },
	"number":   func() any { var t float64; return &t },
	"Selector": func() any { return &Selector{} },
	"Sink":     func() any { return &Sink{} },
	"Source":   func() any { return &Source{} },
	"String":   func() any { return &String{} },
	"string":   func() any { var t string; return &t },
	"Var":      func() any { return &Var{} },
}
