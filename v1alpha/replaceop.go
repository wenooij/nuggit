package v1alpha

import "strconv"

type Repeat struct {
	Min  uint `json:"min,omitempty"`
	Max  uint `json:"max,omitempty"`
	Lazy bool `json:"lazy,omitempty"`
}

//go:generate stringer -type ReplaceOp -linecomment
type ReplaceOp int

const (
	ReplaceUndefined ReplaceOp = iota //
	ReplaceByte                       // byte
)

func (o ReplaceOp) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(o.String())), nil }
