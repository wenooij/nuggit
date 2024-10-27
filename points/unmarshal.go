package points

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/status"
)

func checkType(p nuggit.Point, v any) bool {
	switch scalar := p.Scalar; scalar {
	case "", nuggit.Bytes, nuggit.String: // SQL doesn't discriminate strings and []bytes.
		_, ok := v.([]byte)
		if !ok {
			_, ok = v.(string)
		}
		return ok

	case nuggit.Bool:
		_, ok := v.(bool)
		return ok

	case nuggit.Int64, nuggit.Uint64: // SQL doesn't discriminate int64 and uint64 (and probably others).
		_, ok := v.(int)
		if !ok {
			if _, ok = v.(int64); !ok {
				_, ok = v.(uint64)
			}
		}
		return ok

	case nuggit.Float64:
		_, ok := v.(float64)
		if !ok {
			_, ok = v.(float32)
		}
		return ok

	default:
		return false
	}
}

func unmarshalNewScalarSlice(scalar nuggit.Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", nuggit.Bytes:
		v = [][]byte{}

	case nuggit.String:
		v = []string{}

	case nuggit.Bool:
		v = []bool{}

	case nuggit.Int64:
		v = []int64{}

	case nuggit.Uint64:
		v = []uint64{}

	case nuggit.Float64:
		v = []float64{}

	default:
		return nil, fmt.Errorf("scalar type is not supported (%q): %w", scalar, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func unmarshalNewScalar(scalar nuggit.Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", nuggit.Bytes:
		v = []byte{}

	case nuggit.String:
		v = ""

	case nuggit.Bool:
		v = false

	case nuggit.Int64:
		v = int64(0)

	case nuggit.Uint64:
		v = uint64(0)

	case nuggit.Float64:
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
func unmarshalNew(p nuggit.Point, data []byte) (any, error) {
	if p.Repeated {
		return unmarshalNewScalarSlice(p.Scalar, data)
	}
	return unmarshalNewScalar(p.Scalar, data)
}

// UnmarshalFlat returns an iterator which flattens data and yields individual elements.
//
// This allows multiple points to be extracted with one exchange call.
func UnmarshalFlat(p nuggit.Point, data []byte) iter.Seq2[any, error] {
	v, err := unmarshalNew(p, data)
	if err == nil && checkType(p, v) {
		return func(yield func(any, error) bool) { yield(v, nil) }
	}
	if p.Repeated {
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
