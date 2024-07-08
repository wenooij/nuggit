package ops

import (
	"strings"

	"github.com/wenooij/wire"
)

var concatProto = wire.Struct(map[uint64]wire.Proto[any]{
	1: wire.Any(wire.Seq(wire.String)), // Elems
	2: wire.Any(wire.RawString),        // Sep
})

// Concat takes elements of strings and concantenates them using an optional seperator.
func Concat(r wire.Reader) (wire.SpanElem[string], error) {
	var (
		sb    strings.Builder
		first bool
		sep   string
	)
	msg, err := concatProto.Read(r)
	if err != nil {
		return wire.SpanElem[string]{}, nil
	}
	for _, e := range msg.Elem() {
		switch e.Num() {
		case 1: // Elems
			for _, s := range e.Val().([]string) {
				if !first {
					sb.WriteString(sep)
				}
				sb.WriteString(s)
			}
		case 2: // Sep
			sep = e.Val().(string)
		}
	}
	return wire.MakeString(sb.String()), nil
}
