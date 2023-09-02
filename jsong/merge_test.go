package jsong

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mergeTestCase struct {
	name     string
	dst      any
	src      any
	dstField string
	srcField string
	want     any
	wantErr  bool
}

func (tc mergeTestCase) runTest(t *testing.T) {
	t.Helper()
	got, gotErr := Merge(tc.dst, tc.src, tc.dstField, tc.srcField)
	if gotErr == nil && tc.wantErr {
		t.Errorf("fallbackMerge(%q): want err = true, got err = false", tc.name)
	} else if gotErr != nil && !tc.wantErr {
		t.Errorf("fallbackMerge(%q): want err = false, got err = %v", tc.name, gotErr)
	}
	if diff := cmp.Diff(tc.want, got); diff != "" {
		t.Errorf("Merge(%q): got diff:\n%s", tc.name, diff)
	}
}

func TestMerge(t *testing.T) {
	for _, tc := range []mergeTestCase{{
		name: "empty",
	}, {
		name: "empty object no merge",
		dst:  map[string]any{},
		want: map[string]any{},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.runTest(t)
		})
	}
}
