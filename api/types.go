package api

import "encoding/json"

type Type = string

const (
	TypeUnspecified Type = "" // Same as TypeBytes.
	TypeBytes       Type = "bytes"
	TypeString      Type = "string"
	TypeBool        Type = "bool"
	TypeInt64       Type = "int64"
	TypeUint64      Type = "uin64"
	TypeFloat64     Type = "float64"
	TypeBigInt      Type = "big_int"
	TypeBigFloat    Type = "big_float"
	TypeDOMElement  Type = "dom_element"
	TypeDOMTree     Type = "dom_tree"
)

type TypedValue struct {
	Type  Type            `json:"type,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}
