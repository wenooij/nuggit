package ops

type Literal struct {
	Literal    string `json:"literal,omitempty"`
	IgnoreCase bool   `json:"ignore_case,omitempty"`
}

type Pattern struct {
	Pattern string `json:"pattern,omitempty"`
}
