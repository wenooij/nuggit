package v1alpha

import "fmt"

type Numeric struct {
	Op  NumericOp `json:"op,omitempty"`
	Lhs *Const    `json:"lhs,omitempty"`
	Rhs *Const    `json:"rhs,omitempty"`
}

type NumericOp string

const (
	NumericOpUnknown NumericOp = ""
	NumericOpLhs     NumericOp = "lhs"
	NumericOpRhs     NumericOp = "rhs"
	NumericOpAdd     NumericOp = "add"
)

var numericBiOpMap = map[NumericOp]func(x *Numeric) (any, error){
	NumericOpUnknown: func(x *Numeric) (any, error) { return x.Lhs, nil },
	NumericOpLhs:     func(x *Numeric) (any, error) { return x.Lhs, nil },
	NumericOpRhs:     func(x *Numeric) (any, error) { return x.Rhs, nil },
	NumericOpAdd:     func(x *Numeric) (any, error) { return nil, fmt.Errorf("not implemented") },
}
