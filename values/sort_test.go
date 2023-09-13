package values

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

type sortTestCase struct {
	name        string
	inputValues []any
	wantValues  []any
	wantPanic   bool
}

func (tc sortTestCase) recover(t *testing.T, testName string) {
	t.Helper()
	if err := recover(); err != nil {
		if !tc.wantPanic {
			t.Errorf("%s(%q): recovered from unexpected panic: %v", testName, tc.name, err)
		}
	} else if tc.wantPanic {
		t.Errorf("%s(%q): test case expected a panic but got no panic", testName, tc.name)
	}
}

func (tc sortTestCase) check(t *testing.T, gotValues []any, testName string) {
	t.Helper()
	if !tc.wantPanic {
		if diff := cmp.Diff(tc.wantValues, gotValues); diff != "" {
			t.Errorf("TestSortByKey(%q): got diff:\n%s", tc.name, diff)
		}
	}
}

func (tc sortTestCase) runTest(t *testing.T) {
	t.Helper()
	defer tc.recover(t, "TestSort")

	gotValues := slices.Clone(tc.inputValues)
	Sort(gotValues)
	tc.check(t, gotValues, "TestSort")
}

type sortByKeyTestCase struct {
	sortTestCase
	inputKey string
}

func (tc sortByKeyTestCase) runTest(t *testing.T) {
	t.Helper()
	defer tc.recover(t, "TestSortByKey")

	gotValues := slices.Clone(tc.inputValues)
	SortByKey(gotValues, tc.inputKey)
	tc.check(t, gotValues, "TestSortByKey")
}

func TestSort(t *testing.T) {
	for _, tc := range []sortTestCase{{
		name: "empty",
	}, {
		name:        "strings",
		inputValues: []any{"c", "b", "a"},
		wantValues:  []any{"a", "b", "c"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.runTest(t)
		})
	}
}

func TestSortByKey(t *testing.T) {
	for _, tc := range []sortByKeyTestCase{{
		sortTestCase: sortTestCase{
			name: "empty",
		},
	}, {
		sortTestCase: sortTestCase{
			name:        "int key sorts by slice index",
			inputValues: []any{[]any{"c"}, []any{"b"}, []any{"a"}},
			wantValues:  []any{[]any{"a"}, []any{"b"}, []any{"c"}},
		},
		inputKey: "0",
	}, {
		sortTestCase: sortTestCase{
			name: "field key sorts by field",
			inputValues: []any{
				struct{ V string }{V: "b"},
				struct{ V string }{V: "c"},
				struct{ V string }{V: "a"},
			},
			wantValues: []any{
				struct{ V string }{V: "a"},
				struct{ V string }{V: "b"},
				struct{ V string }{V: "c"},
			},
		},
		inputKey: "V",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.runTest(t)
		})
	}
}
