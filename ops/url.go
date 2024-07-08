package ops

import (
	"net/url"

	"github.com/wenooij/wire"
)

var urlPathEscape = wireFunc(wire.String, wire.String, func(u wire.SpanElem[string]) (wire.SpanElem[string], error) {
	return wire.String.Make(url.PathEscape(u.Elem())), nil
})

func URLPathEscape(r wire.Reader) (wire.Reader, error) { return urlPathEscape(r) }
