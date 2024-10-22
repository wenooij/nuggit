package table

import (
	"fmt"
	"strings"
)

// singleQuote is like strconv.Quote but uses single quotes and two single quote chars as an escape.
func singleQuote(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return fmt.Sprintf("'%s'", s)
}
