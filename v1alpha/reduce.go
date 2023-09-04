package v1alpha

type ReduceOp string

const (
	ReduceUndefined ReduceOp = ""
	ReduceAny       ReduceOp = "any"
	ReduceFirst     ReduceOp = "first"
	ReduceLast      ReduceOp = "last"
	ReduceMin       ReduceOp = "min"
	ReduceMax       ReduceOp = "max"
	ReduceSum       ReduceOp = "sum"
	ReduceVariance  ReduceOp = "var"
	ReduceStddev    ReduceOp = "std"
	ReduceMean      ReduceOp = "mean"
)

type Reduce struct {
	Expr any   `json:"expr,omitempty"`
	Args []any `json:"args,omitempty"`
}
