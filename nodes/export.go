package nodes

import (
	"github.com/wenooij/nuggit/api"
)

type Export struct {
	Nullable        bool       `json:"nullable,omitempty"`
	IncludeMetadata bool       `json:"include_metadata,omitempty"`
	Point           *api.Point `json:"point,omitempty"`
	Cast            *Cast      `json:"cast,omitempty"`
}

type Cast struct {
	From api.Type `json:"from,omitempty"`
	To   api.Type `json:"to,omitempty"`
}
