package api

const runtimesBaseURI = "/api/runtimes"

type Runtime struct {
	Name             string   `json:"name,omitempty"`
	SupportedActions []string `json:"supported_actions,omitempty"`
}

func (r *Runtime) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Runtime) GetSupportedActions() []string {
	if r == nil {
		return nil
	}
	return r.SupportedActions
}
