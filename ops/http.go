package ops

import (
	"io"
	"net/http"

	"github.com/wenooij/wire"
)

var httpGetProto = wire.Span(wire.Seq(wire.Fields(map[uint64]wire.Proto[any]{
	1: wire.Any(wire.RawString), // URL
})))

var httpGetResponseContents = wire.Seq(wire.Fields(map[uint64]wire.Proto[any]{
	1: wire.Any(wire.Raw), // Body
}))

func Get(r wire.Reader) (wire.SpanElem[[]wire.FieldVal[any]], error) {
	msg, err := httpGetProto.Read(r)
	if err != nil {
		return wire.SpanElem[[]wire.FieldVal[any]]{}, nil
	}
	var url string
	for _, e := range msg.Elem() {
		switch e.Num() {
		case 1: // URL
			url = e.Val().(string)
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return wire.SpanElem[[]wire.FieldVal[any]]{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wire.SpanElem[[]wire.FieldVal[any]]{}, err
	}
	return wire.MakeSpan(httpGetResponseContents)([]wire.FieldVal[any]{
		wire.MakeAnyField(wire.Raw)(1, body),
	}), nil
}
