package v1alpha

import (
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/runtime"
)

type Factory struct{}

func (Factory) NewRunner(n nuggit.Node) (runtime.Runner, error) {
	opFn, ok := factoryMapping[n.Op]
	if !ok {
		return nil, fmt.Errorf("v1alpha.Factory.NewRunner: Node %q: Runner is not defined for Op: %q", n.Key, n.Op)
	}
	op := opFn()
	if len(n.Data) > 0 {
		if err := json.Unmarshal(n.Data, op); err != nil {
			return nil, err
		}
	}
	return op, nil
}

var factoryMapping = map[string]func() runtime.Runner{
	"Chromedp": func() runtime.Runner { return &Chromedp{} },
	"Const":    func() runtime.Runner { return &Const{} },
	"HTML":     func() runtime.Runner { return &HTML{} },
	"HTTP":     func() runtime.Runner { return &HTTP{} },
	"Selector": func() runtime.Runner { return &Selector{} },
	"Sink":     func() runtime.Runner { return &Sink{} },
	"Source":   func() runtime.Runner { return &Source{} },
	"String":   func() runtime.Runner { return &String{} },
	"Var":      func() runtime.Runner { return &Var{} },
}
