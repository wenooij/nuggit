package v1alpha

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:generate stringer -type StringOp -linecomment
type StringOp int

const (
	StringUndefined        StringOp = iota //
	StringIdentity                         // identity
	StringAggstring                        // aggstring
	StringIndex                            // index
	StringSubstring                        // substring
	StringToLower                          // tolower
	StringToUpper                          // toupper
	StringURLEncode                        // urlencode
	StringURLDecode                        // urldecode
	StringURLPathEscape                    // urlpathescape
	StringURLPathUnescape                  // urlpathunescape
	StringURLQueryEscape                   // urlqueryescape
	StringURLQueryUnescape                 // urlqueryunescape
)

func (o StringOp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(o.String())), nil
}

// UnmarshalJSON unmarshals the StringOp from a JSON string or integer.
func (o *StringOp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err == nil {
			*o = StringOp(i)
			return nil
		}
		return err
	}
	s = strings.ToLower(s)
	for i := StringOp(0); i < StringOp(len(_StringOp_index)-1); i++ {
		if s == i.String() {
			*o = i
			return nil
		}
	}
	return fmt.Errorf("cannot unmarshal string into StringOp: StringOp not defined for %q", s)
}
