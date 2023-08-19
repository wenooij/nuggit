package v1alpha

import "strconv"

//go:generate stringer -type FnOp -linecomment
type FnOp int

const (
	FnUndefined   FnOp = iota //
	FnPassthrough             // passthrough
	FnFilter                  // filter
	FnMap                     // map
	FnFlatMap                 // flatmap
	FnReduce                  // reduce
	FnHead                    // head
	FnTail                    // tail
)

func (o FnOp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(o.String())), nil
}
