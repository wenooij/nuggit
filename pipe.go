package nuggit

type Pipe struct {
	Actions []Action `json:"actions,omitempty"`
	Point   Point    `json:"point,omitempty"`
}

func (p Pipe) GetSpec() any { return p }
