package table

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type Builder struct {
	c     *api.Collection
	pipes map[api.NameDigest]*api.Pipe
}

func (b *Builder) Reset(c *api.Collection) {
	b.c = c
	b.pipes = make(map[api.NameDigest]*api.Pipe)
}

func (b *Builder) Add(pipes ...*api.Pipe) error {
	for _, p := range pipes {
		if err := b.addPipe(p); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) addPipe(p *api.Pipe) error {
	if p == nil {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	b.pipes[p.NameDigest] = p
	return nil
}

var name = regexp.MustCompile(`^(?i:[a-z][a-z0-9_]*)$`)

func mustValidatedName(s string) string {
	if err := validateName(s); err != nil {
		panic(err)
	}
	return s
}

func validateName(s string) error {
	if s == "" {
		return fmt.Errorf("name is empty: %w", status.ErrInvalidArgument)
	}
	if !name.MatchString(s) {
		return fmt.Errorf("name contains invalid characters (%q): %w", s, status.ErrInvalidArgument)
	}
	return nil
}

func transformName(s string) string { return strings.ReplaceAll(s, "-", "_") }

func (b *Builder) tableName() (string, error) {
	transformed := fmt.Sprintf("collection_%s__%s", transformName(b.c.GetName()), b.c.GetDigest())
	if err := validateName(transformed); err != nil {
		return "", err
	}
	return transformed, nil
}

func (b *Builder) validateBuild() error {
	if b == nil || b.c == nil {
		return fmt.Errorf("table builder is uninitialized: %w", status.ErrInternal)
	}
	// Check table name.
	if _, err := b.tableName(); err != nil {
		return nil
	}
	// Check expected pipes have corresponding pipe objects.
	for _, p := range b.c.GetPipes() {
		pipe, found := b.pipes[p]
		if !found {
			return fmt.Errorf("pipe not found in builder context (%q): %w", p.String(), status.ErrInvalidArgument)
		}
		// Check that the names of pipes conform to naming rules.
		if err := validateName(pipe.GetName()); err != nil {
			return fmt.Errorf("failed validation of pipe in builder context: %w", err)
		}
		// Validate the point (note that the point is already not nil).
		if err := api.ValidatePoint(pipe.Point); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writeColExpr(sb *strings.Builder, name string, point *api.Point) error {
	name = mustValidatedName(name)
	sb.WriteString("    ") // Indent.
	fmt.Fprintf(sb, `%q `, name)
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

func (b *Builder) Build() (string, error) {
	if err := b.validateBuild(); err != nil {
		return "", err
	}

	tableName, err := b.tableName()
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
		if i > 0 {
			sb.WriteString(",\n")
		}
		if err := b.writeColExpr(&sb, name, point); err != nil {
			return "", fmt.Errorf("failed to format column (%q): %w", name, err)
		}
	}
	sb.WriteString("\n);")
	return sb.String(), nil
}

// -- CREATE TABLE
// --     IF NOT EXISTS CollectionData (
// --         ID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
// --         TriggerID TEXT,
// --         CollectionName TEXT NOT NULL,
// --         CollectionDigest TEXT NOT NULL,
// --         DataRow TEXT NOT NULL CHECK (
// --             json_valid (DataRow)
// --             AND json_type (DataRow) = 'array'
// --         ),
// --         FOREIGN KEY (CollectionName) REFERENCES Collections (Name),
// --         FOREIGN KEY (CollectionDigest) REFERENCES Collections (Digest),
// --         FOREIGN KEY (TriggerID) REFERENCES Triggers (TriggerID)
// --     );
