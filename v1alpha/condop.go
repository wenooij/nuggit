package v1alpha

type CondOp string

const (
	CondUndefined   CondOp = ""
	CondPassthrough CondOp = "passthrough"
	CondTrue        CondOp = "true"
	CondEqual       CondOp = "equal"
	CondLess        CondOp = "less"
)
