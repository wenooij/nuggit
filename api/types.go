package api

type ScalarType = string

const (
	TypeUndefined ScalarType = "" // Same as TypeBytes.
	TypeBytes     ScalarType = "bytes"
	TypeString    ScalarType = "string"
	TypeBool      ScalarType = "bool"
	TypeInt64     ScalarType = "int64"
	TypeUint64    ScalarType = "uin64"
	TypeFloat64   ScalarType = "float64"
)

type Type struct {
	Nullable bool       `json:"nullable,omitempty"`
	Repeated bool       `json:"repeated,omitempty"`
	Scalar   ScalarType `json:"scalar,omitempty"`
}

func MakeTypeFromNumber(n int) Type {
	var t Type
	if n&(1<<31) != 0 {
		t.Nullable = true
	}
	if n&(1<<30) != 0 {
		t.Repeated = true
	}
	switch n & 0x7 {
	case 0:

	case 1:
		t.Scalar = TypeString

	case 2:
		t.Scalar = TypeBool

	case 3:
		t.Scalar = TypeInt64

	case 4:
		t.Scalar = TypeUint64

	case 5:
		t.Scalar = TypeFloat64

	default:
	}

	return t
}

func (t Type) Number() int {
	n := 0
	if t.Nullable {
		n |= 1 << 31
	}
	if t.Repeated {
		n |= 1 << 30
	}
	switch t.Scalar {
	case TypeUndefined, TypeBytes:

	case TypeString:
		n |= 1

	case TypeBool:
		n |= 2

	case TypeInt64:
		n |= 3

	case TypeUint64:
		n |= 4

	case TypeFloat64:
		n |= 5

	default:
	}

	return n
}

func scalar(t ScalarType) Type   { return Type{Scalar: t} }
func repeated(t ScalarType) Type { return Type{Repeated: true, Scalar: t} }
