package nuggit

// Type describes supported native types for bootstrapping Nuggit.
// Types unmarshaled from JSON undergo a strict corresion process
// which may result in ErrType is types fail to match.
//
// See Op specific documentation for more Compound types.
type Type string

const (
	TypeUndefined Type = ""
	TypeBool      Type = "bool"
	TypeInt8      Type = "int8"
	TypeInt16     Type = "int16"
	TypeInt32     Type = "int32"
	TypeInt64     Type = "int64"
	TypeUint8     Type = "uint8"
	TypeUint16    Type = "uint16"
	TypeUint32    Type = "uint32"
	TypeUint64    Type = "uint64"
	TypeFloat32   Type = "float32"
	TypeFloat64   Type = "float64"
	TypeBytes     Type = "bytes"
	TypeString    Type = "string"
)
