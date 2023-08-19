package v1alpha

import "strconv"

//go:generate stringer -type NumericOp -linecomment
type NumericOp int

const (
	NumericOpUnknown     NumericOp = iota //
	NumericOpPassthrough                  // passthrough
	NumericOpAdd                          // add
)

func (o NumericOp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(o.String())), nil
}
