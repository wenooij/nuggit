package client

type Args struct {
	URL      string
	Elements map[string][]Element
}

type Element struct {
	Name       string            `json:"name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Text       string            `json:"text,omitempty"`
	Children   []*Element        `json:"children,omitempty"`
}
