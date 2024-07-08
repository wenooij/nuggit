package ops

import (
	"bytes"

	"github.com/wenooij/wire"
)

func wireFunc[T, R any](tp wire.Proto[T], rp wire.Proto[R], f func(T) (R, error)) func(wire.Reader) (wire.Reader, error) {
	var b bytes.Buffer
	return func(r wire.Reader) (wire.Reader, error) {
		t, err := tp.Read(r)
		if err != nil {
			return nil, err
		}
		rv, err := f(t)
		if err != nil {
			return nil, err
		}
		b.Reset()
		b.Grow(int(rp.Size(rv)))
		rp.Write(&b, rv)
		return &b, nil

	}
}
