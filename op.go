package nuggit

// Op is an alias for a string describing an op.
//
// See Node.Op.
type Op = string

const (
	OpUndefined Op = ""
	OpArray     Op = "Array"
	OpAssert    Op = "Assert"
	OpBool      Op = "bool"
	OpCache     Op = "Cache"
	OpChromedp  Op = "Chromedp"
	OpCond      Op = "Cond"
	OpFile      Op = "File"
	OpFind      Op = "Find"
	OpGroupBy   Op = "GroupBy"
	OpHTML      Op = "HTML"
	OpHTTP      Op = "HTTP"
	OpMap       Op = "map"
	OpPoint     Op = "Point"
	OpRegex     Op = "Regex"
	OpReplace   Op = "Replace"
	OpSelect    Op = "Select"
	OpSink      Op = "Sink"
	OpSource    Op = "Source"
	OpSpan      Op = "Span"
	OpString    Op = "string"
	OpTable     Op = "Table"
	OpTime      Op = "Time"
	OpVar       Op = "Var"
)
