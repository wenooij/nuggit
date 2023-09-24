package values

import (
	"fmt"
	"strings"

	"github.com/wenooij/nuggit/jsong"
)

type MapReducer map[string]Reducer

func (a *MapReducer) Add(x any) {
	for k, r := range *a {
		v, err := jsong.Extract(x, k)
		if err != nil {
			panic(fmt.Errorf("MapReducer: failed to extract %q: %w", k, err))
		}
		r.Add(v)
	}
}

func (a *MapReducer) Value() any {
	var m any = make(map[string]any)
	for k, r := range *a {
		mv, err := jsong.Merge(m, r.Value(), k, "")
		if err != nil {
			panic(fmt.Errorf("ReducerMap: failed to merge %q: %w", k, err))
		}
		m = mv
	}
	return m
}

type Reducer interface {
	Add(any)
	Value() any
}

type StringAgg struct {
	strings.Builder
}

func (a *StringAgg) Add(x any) {
	a.Builder.WriteString(x.(string))
}

type ReduceOp string

const (
	ReduceUndefined ReduceOp = ""
	ReduceSum       ReduceOp = "sum"
	ReduceMin       ReduceOp = "min"
	ReduceMax       ReduceOp = "max"
	ReduceAny       ReduceOp = "any"
	ReduceAvg       ReduceOp = "avg"
)

type NumberReducer struct {
	Op  ReduceOp
	val float64
	cnt int
	set bool
}

func (a *NumberReducer) Add(x any) {
	v := x.(float64)
	switch a.Op {
	case ReduceUndefined, ReduceSum:
		a.set = true
		a.val += v
	case ReduceMin:
		if !a.set || v < a.val {
			a.set = true
			a.val = v
		}
	case ReduceMax:
		if !a.set || v > a.val {
			a.set = true
			a.val = v
		}
	case ReduceAny:
		if !a.set {
			a.set = true
			a.val = v
		}
	case ReduceAvg:
		a.set = true
		a.val += v
		a.cnt++
	}
}

func (a *NumberReducer) Value() any {
	if a.Op == ReduceAvg {
		return a.val / float64(a.cnt)
	}
	if !a.set {
		return (*float64)(nil)
	}
	return a.val
}

type SumReducer struct {
	sum float64
}

func (a *SumReducer) Add(x any) {
	a.sum += x.(float64)
}

type TrueCounter struct {
	count int
}

func (a *TrueCounter) Add(x any) {
	if x.(bool) {
		a.count++
	}
}

func (a *StringAgg) Value() any   { return a.String() }
func (a *SumReducer) Value() any  { return a.sum }
func (a *TrueCounter) Value() any { return float64(a.count) }
