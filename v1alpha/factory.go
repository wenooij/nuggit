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
	"Any":      func() any { return new(Any) },
	"Chromedp": func() any { return &Chromedp{} },
	"Const":    func() any { return &Const{} },
	"HTML":     func() any { return &HTML{} },
	"HTTP":     func() any { return &HTTP{} },
	"Selector": func() any { return &Selector{} },
	"Sink":     func() any { return &Sink{} },
	"Source":   func() any { return &Source{} },
	"String":   func() any { return &String{} },
	"Var":      func() any { return &Var{} },
}
