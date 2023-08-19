package nuggit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Glom describes a "glom": a binary operation on inputs.
// Setting the Glom is important when there are multiple possible
// meanings for the data flow between SrcField and DstField.
// The default Glom depends on the specific Ops involved.
// If the Glom is not supportd in a given context ErrGlom should be returned.
//
// Example:
//
//	[A] *assign [B C] = [B C]
//	[A] *append [B C] = [A [B C]]
//	[A] *extend [B C] = [A B C]
//
//go:generate stringer -type Glom -linecomment
type Glom int

const (
	// GlomUndefined applies the default glom operation.
	GlomUndefined Glom = iota //
	// GlomAssign applies a glom which directly assigns DstField to SrcField.
	GlomAssign // assign
	// GlomAppend applies a glom which appends DstField to SrcField.
	GlomAppend // append
	// GlomExtend applies a glom which extends DstField onto SrcField.
	GlomExtend // extend
)

// MarshalJSON marshals the Glom as a JSON string.
func (m Glom) MarshalJSON() ([]byte, error) {
	// TODO(wes): Consider supporting int encoding.
	return []byte(strconv.Quote(m.String())), nil
}

// UnmarshalJSON unmarshals the Glom from a JSON string or integer.
func (m *Glom) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err == nil {
			*m = Glom(i)
			return nil
		}
		return err
	}
	s = strings.ToLower(s)
	for i := Glom(0); i < Glom(len(_Glom_index)-1); i++ {
		if s == i.String() {
			*m = i
			return nil
		}
	}
	return fmt.Errorf("cannot unmarshal string into Glom: Glom not defined for %q", s)
}
