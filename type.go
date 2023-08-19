package nuggit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Type describes supported native types for bootstrapping Nuggit.
// Types unmarshaled from JSON undergo a strict corresion process
// which may result in ErrType is types fail to match.
//
// See Op specific documentation for more Compound types.
//
//go:generate stringer -linecomment -type Type
type Type int

const (
	TypeUndefined Type = iota //
	TypeBool                  // bool
	TypeInt8                  // int8
	TypeInt16                 // int16
	TypeInt32                 // int32
	TypeInt64                 // int64
	TypeUint8                 // uint8
	TypeUint16                // uint16
	TypeUint32                // uint32
	TypeUint64                // uint64
	TypeFloat32               // float32
	TypeFloat64               // float64
	TypeBytes                 // bytes
	TypeString                // string
)

// MarshalJSON marshals the Type as a JSON string.
func (t Type) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(t.String())), nil }

// UnmarshalJSON unmarshals the Type from a JSON string or integer.
func (t *Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err == nil {
			if 0 <= i && i < len(_Type_index)-1 {
				*t = Type(i)
				return nil
			}
			return fmt.Errorf("Type not defined for %v", i)
		}
		return err
	}
	s = strings.ToLower(s)
	for i := Type(0); i < Type(len(_Type_index)-1); i++ {
		if s == i.String() {
			*t = i
			return nil
		}
	}
	return fmt.Errorf("Type not defined for %q", s)
}
