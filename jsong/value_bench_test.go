package jsong

import (
	"testing"
)

var benchValueOfTestCases = []any{
	false,
	0,
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

func BenchmarkValueOf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range benchValueOfTestCases {
			if _, err := ValueOf(v); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkValueOfBaseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range benchValueOfTestCases {
			if _, err := fallbackValueOf(v); err != nil {
				b.Fatal(err)
			}
		}
	}
}
