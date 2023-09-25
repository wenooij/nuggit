package edges

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

// ShortFormat formats the edge including SrcField and DstField, if set.
func Format(e nuggit.Edge) string {
	srcField := e.SrcField
	if srcField == "" {
		srcField = "*"
	}
	dstField := e.DstField
	if dstField == "" {
		dstField = "*"
	}
	var srcGraph, dstGraph string
	if srcGraph != dstGraph {
		srcField = fmt.Sprintf("%q!%s", srcGraph, srcField)
		dstField = fmt.Sprintf("%q!%s", dstGraph, dstField)
	}
	return fmt.Sprintf("%s: %s -> %s", e.Key, srcField, dstField)
}
