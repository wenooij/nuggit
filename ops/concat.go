package ops

import (
	"strings"

	"github.com/wenooij/wire"
)

var concatRequest = map[uint64]wire.Proto[any]{
	1: wire.Any(wire.Seq(wire.String)), // Elems
	2: wire.Any(wire.RawString),        // Sep
}

var concatRequestProto = wire.Struct(concatRequest)

var concat = wireFunc(concatRequestProto, wire.String, func(req wire.SpanElem[[]wire.FieldVal[any]]) (wire.SpanElem[string], error) {
	var (
		sb    strings.Builder
		first bool
		sep   string
	)
	for _, f := range req.Elem() {
		switch f.Num() {
		case 1: // Elems
			for _, s := range f.Val().([]string) {
				if !first {
					sb.WriteString(sep)
				}
				sb.WriteString(s)
			}
		case 2: // Sep
			sep = f.Val().(string)
		}
	}
	return wire.String.Make(sb.String()), nil
})

// Concat takes elements of strings and concantenates them using an optional seperator.
func Concat(r wire.Reader) (wire.Reader, error) { return concat(r) }
