package nuggit

type Rule struct {
	Hostname   string   `json:"hostname,omitempty"`
	URLPattern string   `json:"url_pattern,omitempty"`
	Labels     []string `json:"labels,omitempty"`
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

func (c *Rule) GetLabels() []string {
	if c == nil {
		return nil
	}
	return c.Labels
}
