package points

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

func checkType(p *api.Point, v any) bool {
	switch scalar := p.GetScalar(); scalar {
	case "", api.Bytes, api.String: // SQL doesn't discriminate strings and []bytes.
		_, ok := v.([]byte)
		if !ok {
			_, ok = v.([]string)
		}
		return ok

	case api.Bool:
		_, ok := v.(bool)
		return ok

	case api.Int64, api.Uint64: // SQL doesn't discriminate int64 and uint64 (and probably others).
		_, ok := v.(int)
		if !ok {
			if _, ok = v.(int64); !ok {
				_, ok = v.(uint64)
			}
		}
		return ok

	case api.Float64:
		_, ok := v.(float64)
		if !ok {
			_, ok = v.(float32)
		}
		return ok

	default:
		return false
	}
}

func unmarshalNewScalarSlice(scalar api.Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", api.Bytes:
		v = [][]byte{}

	case api.String:
		v = []string{}

	case api.Bool:
		v = []bool{}

	case api.Int64:
		v = []int64{}

	case api.Uint64:
		v = []uint64{}

	case api.Float64:
		v = []float64{}

	default:
		return nil, fmt.Errorf("scalar type is not supported (%q): %w", scalar, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func unmarshalNewScalar(scalar api.Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", api.Bytes:
		v = []byte{}

	case api.String:
		v = ""

	case api.Bool:
		v = false

	case api.Int64:
		v = int64(0)

	case api.Uint64:
		v = uint64(0)

	case api.Float64:
		v = float64(0)

	default:
		return nil, fmt.Errorf("scalar type is not supported (%q): %w", scalar, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// unmarshalNew is used to convert JSON data from an exchange to point data.
//
// TODO: Currently only !Nullable supported. Handle Nullable.
func unmarshalNew(p *api.Point, data []byte) (any, error) {
	if p.GetRepeated() {
		return unmarshalNewScalarSlice(p.GetScalar(), data)
	}
	return unmarshalNewScalar(p.GetScalar(), data)
}

// UnmarshalFlat returns an iterator which flattens data and yields individual elements.
//
// This allows multiple points to be extracted with one exchange call.
func UnmarshalFlat(p *api.Point, data []byte) iter.Seq2[any, error] {
	v, err := unmarshalNew(p, data)
	if err == nil && checkType(p, v) {
		return func(yield func(any, error) bool) { yield(v, nil) }
	}
	if p.GetRepeated() {
		// TODO: Implement this.
		err := fmt.Errorf("flat unmarshal of repeated values is not yet supported: %w", status.ErrUnimplemented)
		return func(yield func(any, error) bool) { yield(nil, err) }
	}
	// Try to unmarshal it again, but as a Repeated point.
	v, err = unmarshalNew(p.AsRepeated(), data)
	if err != nil {
		return func(yield func(any, error) bool) { yield(nil, err) }
	}

	// v is now a valid slice of type [Scalar].
	// Yield each element of the slice.
	return func(yield func(any, error) bool) {
		v := reflect.ValueOf(v)
		for i := 0; i < v.Len(); i++ {
			if e := v.Index(i); !yield(e.Interface(), nil) {
				return
			}
		}
	}
}
