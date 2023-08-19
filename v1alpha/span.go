package v1alpha

// Span is a representation of an entity as indices into a source byte slice.
//
// The entity content is represented by data[Pos:End] where data is a byte slice.
// The zero entity is used to indicate lack of presence.
// Entities where Pos == End are not valid.
type Span struct {
	Pos int `json:"pos,omitempty"`
	End int `json:"end,omitempty"`
}
