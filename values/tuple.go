package values

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/wenooij/nuggit/jsonglom"
)

type Tuple []any

type ByKey struct {
	Key   string
	Tuple []Tuple
}

func (a ByKey) Len() int      { return len(a.Tuple) }
func (a ByKey) Swap(i, j int) { a.Tuple[i], a.Tuple[j] = a.Tuple[j], a.Tuple[i] }
func (a ByKey) Less(i, j int) bool {
	x, err := jsonglom.Extract(a.Key, a.Tuple[i])
	if err != nil {
		panic(fmt.Errorf("failed to extract key at (%d, _): %w", i, err))
	}
	y, err := jsonglom.Extract(a.Key, a.Tuple[j])
	if err != nil {
		panic(fmt.Errorf("failed to extract key at (_, %d): %w", j, err))
	}
	if k1, k2 := reflect.TypeOf(x).Kind(), reflect.TypeOf(y).Kind(); k1 != k2 {
		panic(fmt.Errorf("mismated Kinds in ByKey sort at (%d, %d): (%v, %v)", i, j, k1, k2))
	}
	switch x := x.(type) {
	case bool:
		return !x && y.(bool)
	case float64:
		return x < y.(float64)
	case []byte:
		return bytes.Compare(x, y.([]byte)) < 0
	case string:
		return x < y.(string)
	default:
		panic(fmt.Errorf("unsupported type in ByKey at (%d, _): %T", i, x))
	}
}
