package values

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

type sortByKeyTestCase struct {
	name        string
	inputTuples [][]any
	inputKey    string
	wantTuples  [][]any
	wantPanic   bool
}

func (tc sortByKeyTestCase) runTest(t *testing.T) {
	t.Helper()

	defer func() {
		if err := recover(); err != nil {
			if !tc.wantPanic {
				t.Errorf("TestSortByKey(%q): recovered from unexpected panic: %v", tc.name, err)
			}
		} else if tc.wantPanic {
			t.Errorf("TestSortByKey(%q): test case expected a panic but got none", tc.name)
		}
	}()

	tuples := slices.Clone(tc.inputTuples)
	SortByKey(tuples, tc.inputKey)

	if !tc.wantPanic {
		if diff := cmp.Diff(tc.wantTuples, tuples); diff != "" {
			t.Errorf("TestSortByKey(%q): got diff:\n%s", tc.name, diff)
		}
	}
}

func TestSortByKey(t *testing.T) {
	for _, tc := range []sortByKeyTestCase{{
		name: "empty",
	}, {
		name:        "empty key",
		inputTuples: [][]any{{"a"}, {"b"}, {"c"}},
		wantPanic:   true,
	}, {
		name:        "index key",
		inputTuples: [][]any{{"c"}, {"b"}, {"a"}},
		inputKey:    "0",
		wantTuples:  [][]any{{"a"}, {"b"}, {"c"}},
	}, {
		name: "index string field",
		inputTuples: [][]any{
			{struct{ V string }{V: "b"}},
			{struct{ V string }{V: "c"}},
			{struct{ V string }{V: "a"}},
		},
		inputKey: "0.V",
		wantTuples: [][]any{
			{struct{ V string }{V: "a"}},
			{struct{ V string }{V: "b"}},
			{struct{ V string }{V: "c"}},
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.runTest(t)
		})
	}
}
