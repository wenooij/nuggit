package nodes

import "github.com/wenooij/nuggit"

type Export struct {
	Nullable        bool                  `json:"nullable,omitempty"`
	IncludeMetadata bool                  `json:"include_metadata,omitempty"`
	Type            nuggit.Type           `json:"type,omitempty"`
	ID              *nuggit.DataSpecifier `json:"id,omitempty"`
	Cast            *Cast                 `json:"cast,omitempty"`
}

type Cast struct {
	From nuggit.Type `json:"from,omitempty"`
	To   nuggit.Type `json:"to,omitempty"`
}
