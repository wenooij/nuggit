package nuggit

// Glom describes a "glom": a binary operation on inputs.
// Setting the Glom is important when there are multiple possible
// meanings for the data flow between SrcField and DstField.
// The default Glom depends on the specific Ops involved.
// If the Glom is not supportd in a given context ErrGlom should be returned.
//
// Example:
//
//	[A] *assign [B C] = [B C]
//	[A] *append [B C] = [A [B C]]
//	[A] *extend [B C] = [A B C]
type Glom string

const (
	// GlomUndefined applies the default glom operation.
	GlomUndefined Glom = ""
	// GlomAssign applies a glom which directly assigns DstField to SrcField.
	GlomAssign Glom = "assign"
	// GlomAppend applies a glom which appends DstField to SrcField.
	GlomAppend Glom = "append"
	// GlomExtend applies a glom which extends DstField onto SrcField.
	GlomExtend Glom = "extend"
)
