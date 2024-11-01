package nuggit

import (
	"strings"
)

type Scalar = string

const (
	Bytes  Scalar = "bytes"
	String Scalar = "string"
	Bool   Scalar = "bool"
	Int    Scalar = "int"
	Float  Scalar = "float"
)

type Point struct {
	Nullable bool   `json:"nullable,omitempty"`
	Scalar   Scalar `json:"scalar,omitempty"`
}

func NewPointFromNumber(x int) Point {
	var p Point
	if x&(1<<31) != 0 {
		p.Nullable = true
	}
	switch x & 0x7 {
	case 0:

	case 1:
		p.Scalar = String

	case 2:
		p.Scalar = Bool

	case 3:
		p.Scalar = Int

	case 4:
		p.Scalar = Float

	default:
	}

	return p
}

func (t Point) AsNumber() int {
	x := 0
	if t.Nullable {
		x |= 1 << 31
	}
	switch t.Scalar {
	case "", Bytes:

	case String:
		x |= 1

	case Bool:
		x |= 2

	case Int:
		x |= 3

	case Float:
		x |= 4

	default:
	}

	return x
}

func (p Point) AsNullable() Point {
	p.Nullable = true
	return p
}

func (p Point) String() string {
	var sb strings.Builder
	sb.Grow(8)
	if p.Nullable {
		sb.WriteByte('*')
	}
	switch p.Scalar {
	case "", Bytes:
		sb.WriteString("bytes")

	case String:
		sb.WriteString("string")

	case Bool:
		sb.WriteString("bytes")

	case Int:
		sb.WriteString("int")

	case Float:
		sb.WriteString("float")

	default:
		sb.WriteString("bytes")
	}
	return sb.String()
}
