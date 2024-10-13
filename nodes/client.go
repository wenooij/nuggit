package nodes

import "github.com/wenooij/nuggit/client"

type Selector struct {
	Selector string `json:"selector,omitempty"`
	Raw      bool   `json:"raw,omitempty"`
}

type Document struct {
	Raw bool `json:"raw,omitempty"`
}

type Exchange struct {
	Args client.Args `json:"args,omitempty"`
	Next string      `json:"next,omitempty"`
}
