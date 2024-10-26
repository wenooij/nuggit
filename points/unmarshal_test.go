package points

import (
	"log"
	"testing"

	"github.com/wenooij/nuggit"
)

func TestUnmarshalFlat(t *testing.T) {
	var p nuggit.Point
	for v, err := range UnmarshalFlat(p, []byte(`["a", "b", "c", "d"]`)) {
		if err != nil {
			t.Fatal(err)
		}
		log.Println(v)
	}
}
