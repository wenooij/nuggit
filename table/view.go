package table

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
)

var name = regexp.MustCompile(`^(?i:[a-z][a-z0-9_]*)$`)

type ViewBuilder struct {
	uuid  string
	alias string

	orderedCols []api.ViewColumn
	pipes       map[integrity.NameDigest]*api.Pipe
	pipeAliases map[integrity.NameDigest]string
}

func (b *ViewBuilder) Reset() {
	b.uuid = ""
	b.alias = ""
	b.orderedCols = make([]api.ViewColumn, 0, 16)
	b.pipes = make(map[integrity.NameDigest]*api.Pipe)
	b.pipeAliases = make(map[integrity.NameDigest]string)
}

func (b *ViewBuilder) SetView(uuid string, alias string) error {
	b.Reset()
	b.uuid = uuid
	b.alias = alias
	return nil
}

func (b *ViewBuilder) AddViewColumn(col api.ViewColumn) error {
	pipe := col.Pipe
	if pipe == nil {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	b.orderedCols = append(b.orderedCols, col)
	b.pipes[pipe.GetNameDigest()] = pipe
	if col.Alias != "" {
		b.pipeAliases[pipe.GetNameDigest()] = col.Alias
	}
	return nil
}

// call transformName first.
func mustValidatedName(s string) string {
	if err := validateName(s); err != nil {
		panic(err)
	}
	return s
}

// call transformName first.
func validateName(s string) error {
	if s == "" {
		return fmt.Errorf("name is empty: %w", status.ErrInvalidArgument)
	}
	if !name.MatchString(s) {
		return fmt.Errorf("name contains invalid characters (%q): %w", s, status.ErrInvalidArgument)
	}
	return nil
}

// transformName idempotently transforms a string to be as a SQL identifier.
//
// WARNING: this still may be an invalid id.
// Use in conjunction with mustValidatedName.
func transformName(s string) string { return strings.ReplaceAll(s, "-", "_") }

func (b *ViewBuilder) viewName() (string, error) {
	transformed := fmt.Sprintf("view_%s", transformName(b.uuid))
	if err := validateName(transformed); err != nil {
		return "", err
	}
	return transformed, nil
}

func (b *ViewBuilder) validateBuild() error {
	if b.uuid == "" {
		return fmt.Errorf("view uuid is empty: %w", status.ErrInternal)
	}
	// Check view name.
	if _, err := b.viewName(); err != nil {
		return nil
	}
	if len(b.orderedCols) == 0 {
		return fmt.Errorf("view must have at least one column: %w", status.ErrInvalidArgument)
	}
	// Check expected pipes have corresponding pipe objects.
	for _, col := range b.orderedCols {
		pipe, found := b.pipes[col.GetNameDigest()]
		if !found {
			return fmt.Errorf("pipe not found in builder context (%q): %w", col.Pipe.String(), status.ErrInvalidArgument)
		}
		// Check that the names of pipes conform to naming rules.
		if err := validateName(transformName(pipe.GetName())); err != nil {
			return fmt.Errorf("failed validation of pipe in builder context: %w", err)
		}
		// Validate the point.
		if err := api.ValidatePoint(pipe.Point); err != nil {
			return err
		}
	}
	if alias := b.alias; alias != "" {
		// Check alias name.
		if err := validateName(transformName(alias)); err != nil {
			return err
		}
	}
	return nil
}

func (b *ViewBuilder) writeSelectColExpr(sb *strings.Builder, col api.ViewColumn) error {
	pipe := col.Pipe
	alias := pipe.GetName()
	if col.Alias != "" {
		alias = col.Alias
	}

	scalarType := "TEXT"
	point := pipe.GetPoint()

	switch scalar := point.Scalar; {
	case point.Repeated: // Array types are simply left as TEXT.
		scalarType = "TEXT"
	default:
		switch scalar {
		case "", nuggit.Bytes:
			scalarType = "BLOB"
		case nuggit.String:
			scalarType = "TEXT"
		case nuggit.Bool:
			scalarType = "BOOLEAN"
		case nuggit.Int64, nuggit.Uint64:
			// There's no UNSIGNED, but we might add a check later in the future.
			scalarType = "INTEGER"
		case nuggit.Float64:
			fmt.Fprint(sb, "REAL")
		default: // Unknown types are simply left as TEXT.
		}
	}

	if err := integrity.ValidateNameDigest(col.Pipe.NameDigest); err != nil {
		return err
	}
	// A valid Pipe name digest is also legal to use in a single quoted string.

	fmt.Fprintf(sb, `MAX (CASE WHEN EXISTS (SELECT 1 FROM Pipes AS p WHERE r.PipeID = p.ID AND p.Name = '%s' AND p.Digest = '%s') THEN CAST(r.Result AS %s) ELSE NULL END) AS %q`,
		col.Pipe.GetName(),
		col.Pipe.GetDigest(),
		scalarType,
		mustValidatedName(transformName(alias)))

	return nil
}

func (b *ViewBuilder) Build() (string, error) {
	if err := b.validateBuild(); err != nil {
		return "", err
	}

	viewName, err := b.viewName()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(256)
	fmt.Fprintf(&sb, "CREATE VIEW IF NOT EXISTS %q AS SELECT\n", mustValidatedName(viewName))

	pipeTableAliases := make(map[integrity.NameDigest]string, len(b.orderedCols))
	for i, col := range b.orderedCols {
		pipeTableAliases[col.GetNameDigest()] = fmt.Sprintf("t%d", i)
	}

	for _, col := range b.orderedCols {
		sb.WriteString("    ") // Indent.
		if err := b.writeSelectColExpr(&sb, col); err != nil {
			return "", fmt.Errorf("failed to format column (%q): %w", name, err)
		}
		sb.WriteString(",\n")
	}

	fmt.Fprintf(&sb, `    e.Timestamp,
    e.URL
FROM TriggerResults AS r
LEFT JOIN TriggerEvents AS e ON r.EventID = e.ID
GROUP BY e.ID, r.SequenceID
ORDER BY e.ID, r.SequenceID ASC;
`)

	// Create view alias.
	if alias := transformName(b.alias); alias != "" {
		fmt.Fprintf(&sb, "CREATE VIEW IF NOT EXISTS %q AS SELECT * FROM %q;\n", mustValidatedName(alias), viewName)
	}

	return sb.String(), nil
}
