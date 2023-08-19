package v1alpha

import "strconv"

//go:generate stringer -type CondOp -linecomment
type CondOp int

const (
	CondUndefined   CondOp = iota //
	CondPassthrough               // passthrough
	CondTrue                      // true
	CondEqual                     // equal
	CondLess                      // less
)

func (m CondOp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(m.String())), nil
}
