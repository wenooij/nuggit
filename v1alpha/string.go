package v1alpha

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wenooij/nuggit/runtime"
)

type String struct {
	String string   `json:"string,omitempty"`
	Op     StringOp `json:"op,omitempty"`
	Sep    string   `json:"sep,omitempty"`
	Begin  int      `json:"begin,omitempty"`
	End    int      `json:"end,omitempty"`
}

func (x *String) Bind(e runtime.Edge) error {
	switch res := e.Result.(type) {
	case string:
		x.String = res
		return nil
	case *Const:
		if s, ok := res.Value.(string); ok {
			x.String = s
		}
		return nil
	default:
		return fmt.Errorf("type error: %T", res)
	}
}

func (x *String) Run(context.Context) (any, error) {
	switch x.Op {
	case StringURLPathEscape:
		return url.PathEscape(x.String), nil
	default:
		return nil, fmt.Errorf("String op undefined for op: %q", x.Op)
	}
}
