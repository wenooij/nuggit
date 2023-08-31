package edges

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

// ShortFormat formats the edge including SrcField and DstField, if set.
func Format(e nuggit.Edge) string {
	if e.SrcField == "" && e.DstField == "" && e.Glom == nuggit.GlomUndefined {
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
	var glomStr string
	glom := e.Glom
	if glom != nuggit.GlomUndefined {
		glomStr = fmt.Sprintf("(%s)", glom)
	}
	return fmt.Sprintf("%s: %s -%s> %s", e.Key, srcField, glomStr, dstField)
}
