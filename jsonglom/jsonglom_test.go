package jsonglom

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wenooij/nuggit"
)

type mergeTestCase struct {
	name          string
	inputData     []byte
	inputField    string
	inputMergeFns []MergeFunc
	want          []byte
	wantErr       bool
}

func (tc mergeTestCase) runTest(t *testing.T) {
	t.Helper()
	got, gotErr := Merge(tc.inputData, tc.inputField, tc.inputMergeFns...)
	if gotErr == nil && tc.wantErr {
		t.Errorf("Merge(%q): want err = true, got err = false", tc.name)
	} else if gotErr != nil && !tc.wantErr {
		t.Errorf("Merge(%q): want err = false, got err = %v", tc.name, gotErr)
	}
	if diff := cmp.Diff(string(tc.want), string(got)); diff != "" {
		t.Errorf("Merge(%q): got diff:\n%s", tc.name, diff)
	}
}

func TestMerge(t *testing.T) {
	for _, tc := range []mergeTestCase{{
		name: "empty input",
		want: []byte(`null`),
	}, {
		name:      "null",
		inputData: []byte(`null`),
		want:      []byte(`null`),
	}, {
		name:      "empty object no merge",
		inputData: []byte(`{}`),
		want:      []byte(`{}`),
	}, {
		name:      "dst is not a JSON object with no merges",
		inputData: []byte(`"abc"`),
		want:      []byte(`"abc"`),
	}, {
		name:      "dst is not a JSON object with merges",
		inputData: []byte(`"abc"`),
		inputMergeFns: []MergeFunc{
			From(json.RawMessage(`{}`), "", nuggit.GlomUndefined),
		},
		wantErr: true,
	}, {
		name:       "simple patch",
		inputData:  []byte(`{}`),
		inputField: "a",
		inputMergeFns: []MergeFunc{
			From(json.RawMessage(`"b"`), "", nuggit.GlomUndefined),
		},
		want: []byte(`{"a":"b"}`),
	}, {
		name:       "simple patch over null",
		inputData:  []byte(`null`),
		inputField: "a",
		inputMergeFns: []MergeFunc{
			From(json.RawMessage(`"b"`), "", nuggit.GlomUndefined),
		},
		want: []byte(`{"a":"b"}`),
	}, {
		name:       "simple patch with 'From' field",
		inputData:  []byte(`{}`),
		inputField: "a",
		inputMergeFns: []MergeFunc{
			From(json.RawMessage(`{"b": 1}`), "b", nuggit.GlomUndefined),
		},
		want: []byte(`{"a":1}`),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.runTest(t)
		})
	}
}
