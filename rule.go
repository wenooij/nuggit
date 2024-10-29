package nuggit

type Rule struct {
	Hostname      string   `json:"hostname,omitempty"`
	URLPattern    string   `json:"url_pattern,omitempty"`
	AlwaysTrigger bool     `json:"always_trigger,omitempty"`
	Disable       bool     `json:"disable,omitempty"`
	Labels        []string `json:"labels,omitempty"`
}

func (c *Rule) GetHostname() string {
	if c == nil {
		return ""
	}
	return c.Hostname
}

func (c *Rule) GetURLPattern() string {
	if c == nil {
		return ""
	}
	return c.URLPattern
}

func (c *Rule) GetAlwaysTrigger() bool {
	if c == nil {
		return false
	}
	return c.AlwaysTrigger
}

func (c *Rule) GetLabels() []string {
	if c == nil {
		return nil
	}
	return c.Labels
}
