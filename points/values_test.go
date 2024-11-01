package points

import (
	"log"
	"testing"

	"github.com/wenooij/nuggit"
)

func TestUnmarshalFlat(t *testing.T) {
	var p nuggit.Point
	p.Scalar = nuggit.Int
	for v, err := range Values(p, []any{1, 2, 3}) {
		if err != nil {
			t.Fatal(err)
		}
		log.Println(v)
	}
}
