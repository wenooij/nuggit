package v1alpha

import "math"

type NumericOp string

const (
	NumericUnknown NumericOp = ""
	NumericZero    NumericOp = "zero"
	NumericOne     NumericOp = "one"
	NumericLeft    NumericOp = "left"
	NumericRight   NumericOp = "right"
	NumericMin     NumericOp = "min"
	NumericMax     NumericOp = "max"
	NumericAdd     NumericOp = "add"
	NumericMul     NumericOp = "mul"
)

var numericOpMap = map[NumericOp]func(x *Numeric, lhs, rhs float64) (float64, error){
	NumericUnknown: func(x *Numeric, lhs, rhs float64) (float64, error) { return float64(0), nil },
	NumericZero:    func(x *Numeric, lhs, rhs float64) (float64, error) { return float64(0), nil },
	NumericOne:     func(x *Numeric, lhs, rhs float64) (float64, error) { return float64(1), nil },
	NumericLeft:    func(x *Numeric, lhs, rhs float64) (float64, error) { return lhs, nil },
	NumericRight:   func(x *Numeric, lhs, rhs float64) (float64, error) { return rhs, nil },
	NumericMin:     func(x *Numeric, lhs, rhs float64) (float64, error) { return math.Min(lhs, rhs), nil },
	NumericMax:     func(x *Numeric, lhs, rhs float64) (float64, error) { return math.Max(lhs, rhs), nil },
	NumericAdd:     func(x *Numeric, lhs, rhs float64) (float64, error) { return lhs + rhs, nil },
	NumericMul:     func(x *Numeric, lhs, rhs float64) (float64, error) { return lhs * rhs, nil },
}

type Numeric struct {
	Op NumericOp `json:"op,omitempty"`
}
