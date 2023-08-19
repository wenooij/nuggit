package v1alpha

import "strconv"

//go:generate stringer -type TimeOp -linecomment
type TimeOp int

const (
	TimeUndefined TimeOp = iota //
	TimeCurrent                 // current
	TimeYear                    // year
)

func (o TimeOp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(o.String())), nil
}
