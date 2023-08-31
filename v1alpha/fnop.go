package v1alpha

type FnOp string

const (
	FnUndefined   FnOp = ""
	FnPassthrough FnOp = "passthrough"
	FnFilter      FnOp = "filter"
	FnMap         FnOp = "map"
	FnFlatMap     FnOp = "flatmap"
	FnReduce      FnOp = "reduce"
	FnHead        FnOp = "head"
	FnTail        FnOp = "tail"
)
