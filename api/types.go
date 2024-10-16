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

func scalar(t ScalarType) Type   { return Type{Scalar: t} }
func repeated(t ScalarType) Type { return Type{Repeated: true, Scalar: t} }
