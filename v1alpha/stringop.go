package v1alpha

import "strconv"

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
