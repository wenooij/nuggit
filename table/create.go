package table

import (
	"fmt"
	"strings"

	"github.com/wenooij/nuggit/api"
)

type CreateBuilder struct {
	Builder
}

func (b *CreateBuilder) writeColExpr(sb *strings.Builder, name string, point *api.Point) error {
	sb.WriteString("    ") // Indent.
	fmt.Fprintf(sb, `%q `, mustValidatedName(transformName(name)))
	repeated := point.GetRepeated()
	var checkUnsigned bool
	switch scalar := point.GetScalar(); {
	case repeated: // JSON array with CHECK.
	default:
		switch scalar {
		case "", api.Bytes:
			fmt.Fprint(sb, "BLOB")
		case api.String:
			fmt.Fprint(sb, "TEXT")
		case api.Bool:
			fmt.Fprint(sb, "BOOLEAN")
		case api.Int64:
			fmt.Fprint(sb, "INTEGER")
		case api.Uint64:
			fmt.Fprint(sb, "INTEGER")
			// There's no UNSIGNED so we add a check.
			// The type is still INTEGER however.
			checkUnsigned = true
		case api.Float64:
			fmt.Fprint(sb, "REAL")
		}
	}
	if !point.GetNullable() {
		sb.WriteString(" NOT NULL")
	}
	if repeated { // Add JSON check.
		fmt.Fprintf(sb, " CHECK (json_valid(%[1]q) AND json_type(%[1]q) = 'array')", name)
	} else if checkUnsigned {
		fmt.Fprintf(sb, " CHECK (%q >= 0)", name)
	}
	return nil
}

func (b *CreateBuilder) Build() (string, error) {
	if err := validateBuild(b.c, b.pipes); err != nil {
		return "", err
	}

	tableName, err := tableName(b.c)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(256)
	fmt.Fprintf(&sb, "CREATE TABLE IF NOT EXISTS %q (\n", tableName)

	for i, p := range b.c.GetPipes() {
		pipe := b.pipes[p]
		name := pipe.GetName()
		point := pipe.GetPoint()
		if err := b.writeColExpr(&sb, name, point); err != nil {
			return "", fmt.Errorf("failed to format column (%q): %w", name, err)
		}
		if i+1 < len(b.c.GetPipes()) {
			sb.WriteByte(',')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString(");")
	return sb.String(), nil
}
