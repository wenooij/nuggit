package nuggit

type Pipe struct {
	Actions []Action `json:"actions,omitempty"`
	Point   Point    `json:"point,omitempty"`
}
