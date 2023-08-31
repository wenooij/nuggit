package v1alpha

import (
	"github.com/wenooij/nuggit"
)

type (
	// Array creates an array of the same record type.
	Array struct {
		Type    nuggit.Type `json:"type,omitempty"`
		Entries []any       `json:"entries,omitempty"`
	}
	// Assert represents an assertion that would cause a program to fail.
	//
	// A string error message can be passed to input.
	Assert struct {
		Lhs *Const `json:"lhs,omitempty"`
		Op  CondOp `json:"op,omitempty"`
		Rhs *Const `json:"rhs,omitempty"`
	}
	Cache struct {
		Dir string `json:"dir,omitempty"`
	}
	Count struct {
		Find *Find `json:"find,omitempty"`
	}
	// CrossJoin uses a lazy cross join strategy to join its arguments.
	CrossJoin struct {
		Lhs any `json:"lhs,omitempty"`
		Rhs any `json:"rhs,omitempty"`
	}
	Entity struct {
		Type  nuggit.Type `json:"type,omitempty"`
		Value string      `json:"value,omitempty"`

		String string `json:"string,omitempty"`
		Table  *Table `json:"table,omitempty"`
	}
	// TODO(wes): Experimental: Determine interface for arbitrary ops and conditions.
	Functional struct {
		Op         FnOp         `json:"op,omitempty"`
		Lambda     []string     `json:"lambda,omitempty"`
		InputType1 *Type        `json:"input_type1,omitempty"`
		InputType2 *Type        `json:"input_type2,omitempty"`
		OutputType *Type        `json:"output_type,omitempty"`
		Input      any          `json:"input,omitempty"`
		Node       *nuggit.Node `json:"node,omitempty"`
	}
	MapEntry struct {
		Key   string
		Value any
	}
	Map struct {
		Data map[string]any `json:"data,omitempty"`
	}
	// Range describes the open interval [Lo, Hi).
	//
	// ExprConfig.Negate may be set to match all values,
	Range struct {
		Lo rune `json:"lo,omitempty"`
		Hi rune `json:"hi,omitempty"`
	}
	// Remote operator specifies a source URL with checksums.
	// Typically used for loading remote Crush programs.
	Remote struct {
		*Source      `json:",omitempty"`
		*nuggit.Sums `json:",omitempty"`
	}
	// Replace replaces strings.
	Replace struct {
		Op ReplaceOp `json:"op,omitempty"`
		// FindSubmatch is the index of the match to find for regex matching.
		FindSubmatch int  `json:"find_submatch,omitempty"`
		FindByte     byte `json:"find_byte,omitempty"`
		ReplaceByte  byte `json:"replace_byte,omitempty"`
		// ReplaceSubmatch replaces the matches with the submatch of the previous regex.
		ReplaceSubmatch int `json:"replace_submatch,omitempty"`
	}
	Row struct{}
	// Sample uses a sampling strategy to select elements from various sources.
	Sample struct{}
)
