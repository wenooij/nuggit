package api

type DOMElement interface {
	domElement()
}

type HTMLElement struct {
	TagName    string            `json:"tag_name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type HTMLElementTree struct {
	TagName    string            `json:"tag_name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []DOMElement      `json:"children,omitempty"`
}

type Args struct {
	Indices []int  `json:"indices,omitempty"`
	String  string `json:"string,omitempty"`
}

type BatchArgs struct {
	*Args        `json:",omitempty"`
	IndicesBatch [][]int  `json:"indices_batch,omitempty"`
	Strings      []string `json:"strings,omitempty"`
}

type Text struct {
	Text string `json:"text,omitempty"`
}

func (HTMLElement) domElement()     {}
func (HTMLElementTree) domElement() {}
func (Text) domElement()            {}
