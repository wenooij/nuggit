package values

import (
	"fmt"
	"sort"

	"github.com/wenooij/nuggit/jsong"
)

func lessVal(a, b any) bool {
	switch a := a.(type) {
	case bool:
		return !a && b.(bool)
	case float64:
		return a < b.(float64)
	case string:
		return a < b.(string)
	default:
		panic(fmt.Errorf("unsupported type in comparison: %T", a))
	}
}

func Sort(vs []any) {
	sort.Slice(vs, func(i, j int) bool { return lessVal(vs[i], vs[j]) })
}

// SortByKey sorts the values by extracting the key using jsong.
func SortByKey(vs []any, key string) {
	if key == "" {
		Sort(vs)
		return
	}
	m := make(map[int]any, len(vs))
	sort.Slice(vs, func(i, j int) bool {
		a, ok := m[i]
		if !ok {
			v, err := jsong.Extract(vs[i], key)
			if err != nil {
				panic(fmt.Errorf("failed to extract key at %d: %w", i, err))
			}
			a = v
		}
		b, ok := m[j]
		if !ok {
			v, err := jsong.Extract(vs[j], key)
			if err != nil {
				panic(fmt.Errorf("failed to extract key at %d: %w", j, err))
			}
			b = v
		}
		return lessVal(a, b)
	})
}
