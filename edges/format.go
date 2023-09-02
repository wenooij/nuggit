package edges

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

// ShortFormat formats the edge including SrcField and DstField, if set.
func Format(e nuggit.Edge) string {
	if e.SrcField == "" && e.DstField == "" {
		return e.Key
	}
	srcField := e.SrcField
	if srcField == "" {
		srcField = "*"
	}
	dstField := e.DstField
	if dstField == "" {
		dstField = "*"
	}
	return fmt.Sprintf("%s: %s -> %s", e.Key, srcField, dstField)
}
