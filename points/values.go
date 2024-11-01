package points

import (
	"fmt"
	"iter"

	"github.com/wenooij/nuggit"
)

func yieldValues[T any](yield func(any, error) bool, data any) bool {
	switch data := data.(type) {
	case T:
		yield(data, nil)
		return true
	case []T:
		for _, e := range data {
			if !yield(e, nil) {
				break
			}
		}
		return true
	case []any:
		if len(data) > 0 {
			if _, ok := data[0].(T); !ok {
				return false
			}
		}
		for _, e := range data {
			if !yield(e, nil) {
				break
			}
		}
		return true
	default:
		return false
	}
}

// Values returns an iterator which flattens data and yields individual elements of the given point type.
func Values(p nuggit.Point, data any) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		switch p.Scalar {
		default: // "", nuggit.Bytes, nuggit.String:
			if !yieldValues[[]byte](yield, data) && !yieldValues[string](yield, data) {
				yield(nil, fmt.Errorf("point value had unexpected type for bytes or string"))
			}
		case nuggit.Bool:
			if !yieldValues[bool](yield, data) {
				yield(nil, fmt.Errorf("point value had unexpected type for bool"))
			}
		case nuggit.Int:
			if !yieldValues[int](yield, data) && !yieldValues[int64](yield, data) {
				yield(nil, fmt.Errorf("point value had unexpected type for int"))
			}
		case nuggit.Float:
			if !yieldValues[float64](yield, data) {
				yield(nil, fmt.Errorf("point value had unexpected type for float"))
			}
		}
	}
}
