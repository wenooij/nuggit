package v1alpha

type Repeat struct {
	Min  uint `json:"min,omitempty"`
	Max  uint `json:"max,omitempty"`
	Lazy bool `json:"lazy,omitempty"`
}

type ReplaceOp string

const (
	ReplaceUndefined ReplaceOp = ""
	ReplaceByte      ReplaceOp = "byte"
)
