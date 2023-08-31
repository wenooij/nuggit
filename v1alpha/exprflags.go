package v1alpha

import (
	"encoding/json"
	"strings"
)

type ExprFlags int

const (
	FlagsUndefined       ExprFlags = 0               //
	FlagsCaseInsensitive ExprFlags = 1 << (iota - 1) // i
	FlagsMultiLine                                   // m
	FlagsDotAll                                      // s
)

func (flags ExprFlags) CaseInsensitive() bool {
	return flags&FlagsCaseInsensitive != 0
}

func (flags ExprFlags) MultiLine() bool {
	return flags&FlagsMultiLine != 0
}

func (flags ExprFlags) DotAll() bool {
	return flags&FlagsDotAll != 0
}

func (flags ExprFlags) String() string {
	var sb strings.Builder
	if flags.CaseInsensitive() {
		sb.WriteByte('i')
	}
	if flags.MultiLine() {
		sb.WriteByte('m')
	}
	if flags.DotAll() {
		sb.WriteByte('s')
	}
	return sb.String()
}

func (flags ExprFlags) MarshalJSON() ([]byte, error) {
	return json.Marshal(flags.String())
}

func (flags *ExprFlags) UnmarshalJSON(data []byte) error {
	var fs ExprFlags
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// TODO(wes): Make stricter.
	if strings.IndexByte(s, 'i') != -1 {
		fs |= FlagsCaseInsensitive
	}
	if strings.IndexByte(s, 'm') != -1 {
		fs |= FlagsMultiLine
	}
	if strings.IndexByte(s, 's') != -1 {
		fs |= FlagsDotAll
	}
	*flags = fs
	return nil
}
