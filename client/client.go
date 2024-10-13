package client

type Args struct {
	URL      string               `json:"url,omitempty"`
	Elements map[string][]Element `json:"elements,omitempty"`
}

const CData = "CDATA"

type Element struct {
	Name       string            `json:"name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Text       string            `json:"text,omitempty"`
	Children   []*Element        `json:"children,omitempty"`
}
