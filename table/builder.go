package table

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

var name = regexp.MustCompile(`^(?i:[a-z][a-z0-9_]*)$`)

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

// transformName idempotently transforms a string to be as a SQL identifier.
//
// WARNING: this still may be an invalid id.
// Use in conjunction with mustValidatedName.
func transformName(s string) string { return strings.ReplaceAll(s, "-", "_") }

func tableName(c *api.Collection) (string, error) {
	transformed := fmt.Sprintf("collection_%s__%s", transformName(c.GetName()), c.GetDigest())
	if err := validateName(transformed); err != nil {
		return "", err
	}
	return transformed, nil
}

func validateBuild(c *api.Collection, pipes map[api.NameDigest]*api.Pipe) error {
	if c == nil {
		return fmt.Errorf("table builder is uninitialized: %w", status.ErrInternal)
	}
	// Check table name.
	if _, err := tableName(c); err != nil {
		return nil
	}
	// Check expected pipes have corresponding pipe objects.
	for _, p := range c.GetPipes() {
		pipe, found := pipes[p]
		if !found {
			return fmt.Errorf("pipe not found in builder context (%q): %w", p.String(), status.ErrInvalidArgument)
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
	return nil
}
