package v1alpha

//go:generate stringer -linecomment -type ExprFlags
type ExprFlags int

const (
	ExprFlagsNone            ExprFlags = 0               //
	ExprFlagsCaseInsensitive ExprFlags = 1 << (iota - 1) // i
	ExprFlagsMultiLine                                   // m
	ExprFlagsDotAll                                      // s
)

func (flags ExprFlags) CaseInsensitive() bool {
	return flags&ExprFlagsCaseInsensitive != 0
}

func (flags ExprFlags) MultiLine() bool {
	return flags&ExprFlagsMultiLine != 0
}

func (flags ExprFlags) DotAll() bool {
	return flags&ExprFlagsDotAll != 0
}

// TODO(wes): Add JSON methods.
