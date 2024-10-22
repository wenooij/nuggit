package api

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"

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

func (p *Point) GetNullable() bool {
	if p == nil {
		return false
	}
	return p.Nullable
}
func (p *Point) GetRepeated() bool {
	if p == nil {
		return false
	}
	return p.Repeated
}
func (p *Point) GetScalar() Scalar {
	if p == nil {
		return ""
	}
	return p.Scalar
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

func (p *Point) AsNullable() *Point {
	var t Point
	if p != nil {
		t = *p
	}
	t.Nullable = true
	return &t
}

func (p *Point) AsScalar() *Point {
	var t Point
	if p != nil {
		t = *p
	}
	t.Repeated = false
	return &t
}

func (p *Point) AsRepeated() *Point {
	var t Point
	if p != nil {
		t = *p
	}
	t.Repeated = true
	return &t
}

func ValidatePoint(p *Point) error {
	// Nil points are allowed and equivalent to the zero point.
	if p == nil {
		return nil
	}
	return ValidateScalar(p.Scalar)
}

func (p *Point) checkType(v any) bool {
	switch scalar := p.GetScalar(); scalar {
	case "", Bytes, String: // SQL doesn't discriminate strings and []bytes.
		_, ok := v.([]byte)
		if !ok {
			_, ok = v.([]string)
		}
		return ok

	case Bool:
		_, ok := v.(bool)
		return ok

	case Int64, Uint64: // SQL doesn't discriminate int64 and uint64 (and probably others).
		_, ok := v.(int)
		if !ok {
			if _, ok = v.(int64); !ok {
				_, ok = v.(uint64)
			}
		}
		return ok

	case Float64:
		_, ok := v.(float64)
		if !ok {
			_, ok = v.(float32)
		}
		return ok

	default:
		return false
	}
}

func unmarshalNewScalarSlice(scalar Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", Bytes:
		v = [][]byte{}

	case String:
		v = []string{}

	case Bool:
		v = []bool{}

	case Int64:
		v = []int64{}

	case Uint64:
		v = []uint64{}

	case Float64:
		v = []float64{}

	default:
		return nil, fmt.Errorf("scalar type is not supported (%q): %w", scalar, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func unmarshalNewScalar(scalar Scalar, data []byte) (any, error) {
	var v any
	switch scalar {
	case "", Bytes:
		v = []byte{}

	case String:
		v = ""

	case Bool:
		v = false

	case Int64:
		v = int64(0)

	case Uint64:
		v = uint64(0)

	case Float64:
		v = float64(0)

	default:
		return nil, fmt.Errorf("scalar type is not supported (%q): %w", scalar, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// unmarshalNew is used to convert JSON data from an exchange to point data.
//
// TODO: Currently only !Nullable supported. Handle Nullable.
func (p *Point) unmarshalNew(data []byte) (any, error) {
	if p.GetRepeated() {
		return unmarshalNewScalarSlice(p.GetScalar(), data)
	}
	return unmarshalNewScalar(p.GetScalar(), data)
}

// UnmarshalFlat returns an iterator which flattens data and yields individual elements.
//
// This allows multiple points to be extracted with one exchange call.
func (p *Point) UnmarshalFlat(data []byte) iter.Seq2[any, error] {
	v, err := p.unmarshalNew(data)
	if err == nil && p.checkType(v) {
		return func(yield func(any, error) bool) { yield(v, nil) }
	}
	if p.GetRepeated() {
		// TODO: Implement this.
		err := fmt.Errorf("flat unmarshal of repeated values is not yet supported: %w", status.ErrUnimplemented)
		return func(yield func(any, error) bool) { yield(nil, err) }
	}
	// Try to unmarshal it again, but as a Repeated point.
	v, err = p.AsRepeated().unmarshalNew(data)
	if err != nil {
		return func(yield func(any, error) bool) { yield(nil, err) }
	}

	// v is now a valid slice of type [Scalar].
	// Yield each element of the slice.
	return func(yield func(any, error) bool) {
		v := reflect.ValueOf(v)
		for i := 0; i < v.Len(); i++ {
			if e := v.Index(i); !yield(e.Interface(), nil) {
				return
			}
		}
	}
}
