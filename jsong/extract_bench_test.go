package jsong

import (
	"testing"
)

var benchExtractValues = []any{
	"",
	"abc",
	[]int{1, 2, 3},
	map[string]int{"a": 1, "b": 2, "c": 3},
	struct {
		X int
		Y struct{ Z *int }
	}{X: 5, Y: struct{ Z *int }{}},
	make([]int, 100),
	make([]byte, 20),
	string(make([]byte, 20)),
	make([]struct{ X, Y, Z **int }, 20),
}

var benchExtractPaths = []string{
	"",
	"0",
	"a",
	"y.z",
	"15",
	"z",
}

func BenchmarkExtract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range benchExtractValues {
			for _, path := range benchExtractPaths {
				Extract(v, path)
			}
		}
	}
}

func BenchmarkExtractBaseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range benchExtractValues {
			for _, path := range benchExtractPaths {
				fallbackExtract(v, path)
			}
		}
	}
}
