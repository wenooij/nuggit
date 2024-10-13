package runtime

type nodeState struct {
	Disabled    bool `json:"disabled,omitempty"`
	Passthrough bool `json:"passthrough,omitempty"`
}

type pipelineState struct {
	Disabled bool `json:"disabled,omitempty"`
}
