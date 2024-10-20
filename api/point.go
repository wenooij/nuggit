package api

import (
	"fmt"

	"github.com/wenooij/nuggit/status"
)

type Scalar = string

const (
	Bytes   Scalar = "bytes"
	String  Scalar = "string"
	Bool    Scalar = "bool"
	Int64   Scalar = "int64"
	Uint64  Scalar = "uin64"
	Float64 Scalar = "float64"
)

var supportedScalars = map[Scalar]struct{}{
	"":      {}, // Same as Bytes.
	Bytes:   {},
	String:  {},
	Bool:    {},
	Int64:   {},
	Uint64:  {},
	Float64: {},
}

func ValidateScalar(s Scalar) error {
	_, ok := supportedScalars[s]
	if !ok {
		return fmt.Errorf("scalar type is not supported (%q): %w", s, status.ErrInvalidArgument)
	}
	return nil
}

type Point struct {
	Nullable bool   `json:"nullable,omitempty"`
	Repeated bool   `json:"repeated,omitempty"`
	Scalar   Scalar `json:"scalar,omitempty"`
}

func NewPointFromNumber(x int) Point {
	var p Point
	if x&(1<<31) != 0 {
		p.Nullable = true
	}
	if x&(1<<30) != 0 {
		p.Repeated = true
	}
	switch x & 0x7 {
	case 0:

	case 1:
		p.Scalar = String

	case 2:
		p.Scalar = Bool

	case 3:
		p.Scalar = Int64

	case 4:
		p.Scalar = Uint64

	case 5:
		p.Scalar = Float64

	default:
	}

	return p
}

func (t Point) AsNumber() int {
	x := 0
	if t.Nullable {
		x |= 1 << 31
	}
	if t.Repeated {
		x |= 1 << 30
	}
	switch t.Scalar {
	case "", Bytes:

	case String:
		x |= 1

	case Bool:
		x |= 2

	case Int64:
		x |= 3

	case Uint64:
		x |= 4

	case Float64:
		x |= 5

	default:
	}

	return x
}

func (p Point) AsNullable() Point {
	p.Nullable = true
	return p
}

func (p Point) AsScalar() Point {
	p.Repeated = false
	return p
}

func (p Point) AsRepeated() Point {
	p.Repeated = true
	return p
}

func ValidatePoint(p *Point) error {
	// Nil points are allowed and equivalent to the zero point.
	if p == nil {
		return nil
	}
	return ValidateScalar(p.Scalar)
}
