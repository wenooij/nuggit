package ops

import "github.com/wenooij/nuggit"

type Cast struct {
	Src    nuggit.Type `json:"src,omitempty"`
	Target nuggit.Type `json:"target,omitempty"`
}
