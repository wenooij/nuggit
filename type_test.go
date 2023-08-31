package nuggit

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMarshalType(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   Type
		want    []byte
		wantErr bool
	}{
		{name: "Undefined", input: TypeUndefined, want: []byte(`""`)},
		{name: "Bool", input: TypeBool, want: []byte(`"bool"`)},
		{name: "Int8", input: TypeInt8, want: []byte(`"int8"`)},
		{name: "Int16", input: TypeInt16, want: []byte(`"int16"`)},
		{name: "Int32", input: TypeInt32, want: []byte(`"int32"`)},
		{name: "Int64", input: TypeInt64, want: []byte(`"int64"`)},
		{name: "Uint8", input: TypeUint8, want: []byte(`"uint8"`)},
		{name: "Uint16", input: TypeUint16, want: []byte(`"uint16"`)},
		{name: "Uint32", input: TypeUint32, want: []byte(`"uint32"`)},
		{name: "Uint64", input: TypeUint64, want: []byte(`"uint64"`)},
		{name: "Float32", input: TypeFloat32, want: []byte(`"float32"`)},
		{name: "Float64", input: TypeFloat64, want: []byte(`"float64"`)},
		{name: "Bytes", input: TypeBytes, want: []byte(`"bytes"`)},
		{name: "String", input: TypeString, want: []byte(`"string"`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := json.Marshal(tc.input)
			if tc.wantErr && gotErr == nil {
				t.Errorf("TestMarshalType(%q): wanted err, got err = nil", tc.name)
			}
			if gotErr != nil {
				t.Log(gotErr)
				if !tc.wantErr {
					t.Errorf("TestMarshalType(%q): wanted err = nil, got err = %v", tc.name, gotErr)
				}
			}
			if diff := cmp.Diff(string(tc.want), string(got)); diff != "" {
				t.Errorf("TestMarshalType(%q): got diff (-want, +got):\n%v", tc.name, diff)
			}
		})
	}
}

func TestUnmarshalType(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   []byte
		want    Type
		wantErr bool
	}{
		{name: "Undefined", input: []byte(`""`), want: TypeUndefined},
		{name: "Bool", input: []byte(`"bool"`), want: TypeBool},
		{name: "Int8", input: []byte(`"int8"`), want: TypeInt8},
		{name: "Int16", input: []byte(`"int16"`), want: TypeInt16},
		{name: "Int32", input: []byte(`"int32"`), want: TypeInt32},
		{name: "Int64", input: []byte(`"int64"`), want: TypeInt64},
		{name: "Uint8", input: []byte(`"uint8"`), want: TypeUint8},
		{name: "Uint16", input: []byte(`"uint16"`), want: TypeUint16},
		{name: "Uint32", input: []byte(`"uint32"`), want: TypeUint32},
		{name: "Uint64", input: []byte(`"uint64"`), want: TypeUint64},
		{name: "Float32", input: []byte(`"float32"`), want: TypeFloat32},
		{name: "Float64", input: []byte(`"float64"`), want: TypeFloat64},
		{name: "Bytes", input: []byte(`"bytes"`), want: TypeBytes},
		{name: "String", input: []byte(`"string"`), want: TypeString},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got Type
			gotErr := json.Unmarshal(tc.input, &got)
			if tc.wantErr && gotErr == nil {
				t.Errorf("TestUnmarshalType(%q): wanted err, got err = nil", tc.name)
			}
			if gotErr != nil {
				t.Log(gotErr)
				if !tc.wantErr {
					t.Errorf("TestUnmarshalType(%q): wanted err = nil, got err = %v", tc.name, gotErr)
				}
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("TestUnmarshalType(%q): got diff (-want, +got):\n%v", tc.name, diff)
			}
		})
	}
}
