package v1alpha

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/exp/utf8string"
)

type String struct {
	Op     StringOp `json:"op,omitempty"`
	String string   `json:"value,omitempty"`
	Format string   `json:"format,omitempty"`
	Sink   *Sink    `json:"sink,omitempty"`
	Args   []any    `json:"args,omitempty"`
	Const  *Const   `json:"const,omitempty"`
	Sep    string   `json:"sep,omitempty"`
	At     int      `json:"at,omitempty"`
	Begin  int      `json:"begin,omitempty"`
	End    int      `json:"end,omitempty"`
}

func (x *String) Run(context.Context) (any, error) {
	s := x.String
	if x.Const != nil {
		s = x.Const.Value.(string)
	}
	x.String = s

	opFn, ok := stringOpMap[x.Op]
	if !ok {
		return nil, fmt.Errorf("String op undefined for op: %q", x.Op)
	}
	return opFn(x)
}

type StringOp string

const (
	StringUndefined        StringOp = ""
	StringFormat           StringOp = "sprintf"
	StringAggstring        StringOp = "aggstring"
	StringSubstring        StringOp = "substring"
	StringToLower          StringOp = "tolower"
	StringToUpper          StringOp = "toupper"
	StringURLPathEscape    StringOp = "urlpathescape"
	StringURLPathJoin      StringOp = "urlpathjoin"
	StringURLPathUnescape  StringOp = "urlpathunescape"
	StringURLQueryEscape   StringOp = "urlqueryescape"
	StringURLQueryUnescape StringOp = "urlqueryunescape"
)

var stringOpMap = map[StringOp]func(*String) (any, error){
	StringUndefined: func(x *String) (any, error) { return fmt.Sprintf(x.Format, x.Args...), nil },
	StringFormat:    func(x *String) (any, error) { return fmt.Sprintf(x.Format, x.Args...), nil },
	StringAggstring: func(x *String) (any, error) {
		var sb strings.Builder
		for i, a := range x.Args {
			if i > 0 {
				sb.WriteString(x.Sep)
			}
			fmt.Fprint(&sb, a)
		}
		return sb.String(), nil
	},
	StringSubstring: func(x *String) (v any, err error) {
		defer func() {
			if err1 := recover(); err1 != nil {
				err = err1.(error)
			}
		}()
		return utf8string.NewString(x.String).Slice(x.Begin, x.End), nil
	},
	StringToLower:       func(x *String) (any, error) { return strings.ToLower(x.String), nil },
	StringToUpper:       func(x *String) (any, error) { return strings.ToUpper(x.String), nil },
	StringURLPathEscape: func(x *String) (any, error) { return url.PathEscape(x.String), nil },
	StringURLPathJoin: func(x *String) (any, error) {
		elem := make([]string, 0, len(x.Args))
		for _, a := range x.Args {
			if s, ok := a.(string); ok {
				elem = append(elem, s)
			} else {
				elem = append(elem, fmt.Sprint(a))
			}
		}
		return url.JoinPath(x.String, elem...)
	},
	StringURLPathUnescape:  func(x *String) (any, error) { return url.PathUnescape(x.String) },
	StringURLQueryEscape:   func(x *String) (any, error) { return url.QueryEscape(x.String), nil },
	StringURLQueryUnescape: func(x *String) (any, error) { return url.QueryUnescape(x.String) },
}
