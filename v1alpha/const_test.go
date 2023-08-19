package v1alpha

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wenooij/nuggit"
)

func TestConstUnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   []byte
		want    Const
		wantErr bool
	}{
		{name: "empty input", wantErr: true},
		{name: "Bool", input: []byte("false"), want: Const{Type: nuggit.TypeBool, Value: false}},
		{name: "Int8", input: []byte("-1"), want: Const{Type: nuggit.TypeInt8, Value: int8(-1)}},
		{name: "Int16", input: []byte("-32768"), want: Const{Type: nuggit.TypeInt16, Value: int16(-32768)}},
		{name: "Int32", input: []byte("-2147483648"), want: Const{Type: nuggit.TypeInt32, Value: int32(-2147483648)}},
		{name: "Int64", input: []byte("-9223372036854775808"), want: Const{Type: nuggit.TypeInt64, Value: int64(-9223372036854775808)}},
		{name: "Uint8", input: []byte("0"), want: Const{Type: nuggit.TypeUint8, Value: uint8(0)}},
		{name: "Uint16", input: []byte("65535"), want: Const{Type: nuggit.TypeUint16, Value: uint16(65535)}},
		{name: "Uint32", input: []byte("4294967295"), want: Const{Type: nuggit.TypeUint32, Value: uint32(4294967295)}},
		{name: "Uint64", input: []byte("18446744073709551615"), want: Const{Type: nuggit.TypeUint64, Value: uint64(18446744073709551615)}},
		{name: "Float64", input: []byte("1.79769313486231570814527423731704356798070e+308"), want: Const{Type: nuggit.TypeFloat64, Value: float64(1.79769313486231570814527423731704356798070e+308)}},
		{name: "String", input: []byte(`"abc"`), want: Const{Type: nuggit.TypeString, Value: "abc"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got Const
			gotErr := json.Unmarshal(tc.input, &got)
			if tc.wantErr && gotErr == nil {
				t.Errorf("TestConstUnmarshalJSON(%q): wanted err, got err = nil", tc.name)
			}
			if gotErr != nil {
				t.Log(gotErr)
				if !tc.wantErr {
					t.Errorf("TestConstUnmarshalJSON(%q): wanted err = nil, got err = %v", tc.name, gotErr)
				}
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("TestConstUnmarshalJSON(%q): got diff (-want, +got):\n%v", tc.name, diff)
			}
		})
	}
}
