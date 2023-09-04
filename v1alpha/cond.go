package v1alpha

type CondOp string

const (
	CondUndefined CondOp = ""
	CondNop       CondOp = "nop"
	CondTrue      CondOp = "true"
	CondFalse     CondOp = "false"
	CondEqual     CondOp = "equal"
	CondLess      CondOp = "less"
	CondNot       CondOp = "not"
)

type Cond struct {
	Op any `json:"op,omitempty"`
}

type OneOf struct {
	Terms []any `json:"terms,omitempty"`
}

type AnyOf struct {
	Terms []any `json:"terms,omitempty"`
}
