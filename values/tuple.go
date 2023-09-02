package values

import (
	"fmt"
	"sort"

	"github.com/wenooij/nuggit/jsong"
)

// SortByKey sorts the tuples by extracting the key using jsonglom.
func SortByKey(tuples [][]any, key string) {
	vs := make([]any, len(tuples))
	for i, t := range tuples {
		v, err := jsong.Extract(t, key)
		if err != nil {
			panic(fmt.Errorf("failed to extract key at %d: %w", i, err))
		}
		vs[i] = v
	}
	fmt.Println(vs)
	sort.Slice(tuples, func(i, j int) bool {
		a, b := vs[i], vs[j]
		switch a := a.(type) {
		case bool:
			return !a && b.(bool)
		case float64:
			return a < b.(float64)
		case string:
			return a < b.(string)
		case nil:
			return false
		default:
			panic(fmt.Errorf("unsupported types in sort at %d: %T", i, a))
		}
	})
	fmt.Println("sorted:", tuples)
}
