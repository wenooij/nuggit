package table

import (
	"fmt"
	"strings"
)

type InsertBuilder struct {
	Builder
}

// n placeholders: ?,?,?,...
func placeholders(n int) string {
	var sb strings.Builder
	sb.Grow(2 * n)
	for ; n > 1; n-- {
		sb.WriteString("?,")
	}
	if n == 1 {
		sb.WriteByte('?')
	}
	return sb.String()
}

func (b *InsertBuilder) Build() (string, error) {
	if err := validateBuild(b.c, b.pipes); err != nil {
		return "", err
	}
	tableName, err := tableName(b.c)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(256)
	fmt.Fprintf(&sb, "INSERT INTO %q (\n", tableName)
	n := len(b.c.GetPipes())
	for i, p := range b.c.GetPipes() {
		pipe := b.pipes[p]
		fmt.Fprintf(&sb, "    %q", pipe.GetName())
		if i+1 < n {
			sb.WriteByte(',')
		}
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, ") VALUES (%s) ON CONFLICT DO NOTHING;\n", placeholders(n))
	return sb.String(), nil
}
