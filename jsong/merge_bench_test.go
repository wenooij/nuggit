package jsong

import (
	"testing"
)

func benchMergeTestCases() []mergeTestCase {
	return []mergeTestCase{{
		dst:      &struct{ X int }{1},
		src:      &struct{ X int }{1},
		dstField: "X",
		srcField: "X",
	}, {
		dst:      []int{1, 2, 3},
		src:      &struct{ X int }{1},
		dstField: "1",
		srcField: "X",
	}, {
		dst:      &struct{ X struct{ Y int } }{},
		src:      []int{1, 2, 3},
		dstField: "x.y",
		srcField: "2",
	}}
}

func BenchmarkMerge(b *testing.B) {
	testCases := benchMergeTestCases()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			Merge(tc.dst, tc.src, tc.dstField, tc.srcField)
		}
	}
}
