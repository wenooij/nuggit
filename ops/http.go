package ops

import (
	"io"
	"net/http"

	"github.com/wenooij/wire"
)

var httpGetRequest = map[uint64]wire.Proto[any]{
	1: wire.Any(wire.RawString), // URL
}

var httpGetResponse = map[uint64]wire.Proto[any]{
	1: wire.Any(wire.Raw), // Body
}

var httpGetRequestProto = wire.Struct(httpGetRequest)

var httpGetResponseProto = wire.Struct(httpGetResponse)

var httpGet = wireFunc(httpGetRequestProto, httpGetResponseProto, func(req wire.SpanElem[[]wire.FieldVal[any]]) (wire.SpanElem[[]wire.FieldVal[any]], error) {
	var url string
	for _, e := range req.Elem() {
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
	s := httpGetResponseProto.Make([]wire.FieldVal[any]{
		wire.Field(wire.Raw).Make(wire.Tup2Val[uint64, []byte]{E0: 1, E1: body}).Any(),
	})
	return s, nil
})

func HTTPGet(r wire.Reader) (wire.Reader, error) { return httpGet(r) }
