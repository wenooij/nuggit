package api

import (
	"log"
	"testing"
)

func TestUnmarshalFlat(t *testing.T) {
	var p *Point
	for v, err := range p.UnmarshalFlat([]byte(`["a", "b", "c", "d"]`)) {
		if err != nil {
			t.Fatal(err)
		}
		log.Println(v)
	}
}
