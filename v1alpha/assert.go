package v1alpha

// Assert represents an assertion that would cause a program to fail.
// A string error message can be passed to input.
type Assert struct {
	Op   any   `json:"op,omitempty"`
	Args []any `json:"rhs,omitempty"`
}
