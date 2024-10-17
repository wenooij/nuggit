package api

type RuntimeLite struct {
	*Ref `json:",omitempty"`
}

func (r *RuntimeLite) GetRef() *Ref {
	if r == nil {
		return nil
	}
	return r.Ref
}

func NewRuntimeLite(id string) *RuntimeLite {
	return &RuntimeLite{newRef("/api/runtimes/", id)}
}

type RuntimeBase struct {
	Name             string   `json:"name,omitempty"`
	SupportedActions []string `json:"supported_actions,omitempty"`
}

type Runtime struct {
	*RuntimeLite `json:",omitempty"`
	*RuntimeBase `json:",omitempty"`
}
