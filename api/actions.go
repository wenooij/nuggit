package api

import (
	"encoding/json"
)

type ActionLite struct {
	*Ref `json:",omitempty"`
}

func NewActionLite(id string) *ActionLite {
	return &ActionLite{newRef("/api/actions/", id)}
}

func (a *ActionLite) UUID() string {
	if a == nil {
		return ""
	}
	return a.Ref.UUID()
}

type ActionBase struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

func (a *ActionBase) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *ActionBase) GetSpec() json.RawMessage {
	if a == nil {
		return nil
	}
	return a.Spec
}

type Action struct {
	*ActionLite `json:",omitempty"`
	*ActionBase `json:",omitempty"`
}

func (a *Action) GetLite() *ActionLite {
	if a == nil {
		return nil
	}
	return a.ActionLite
}

func (a *Action) GetBase() *ActionBase {
	if a == nil {
		return nil
	}
	return a.ActionBase
}

const (
	ActionAttribute = "attribute" // AttributeAction extracts attribute names from the
	ActionDocument  = "document"  // ActionDocument represents an action which copies the full document.
	ActionExchange  = "exchange"  // ActionExchange marks the (network) boundary for a client-server data exchange.
	ActionPattern   = "pattern"   // ActionPattern matches re2 patterns.
	ActionSelector  = "selector"  // ActionSelector matches CSS selectors.
)

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
}

type DocumentAction struct{}

type AttributeAction struct {
	Attribute string `json:"attribute,omitempty"`
}

type ExchangeAction struct {
	Pipe string `json:"pipe,omitempty"`
}

type PatternActionBase struct {
	Pattern         string `json:"pattern,omitempty"`
	Passthrough     bool   `json:"passthrough,omitempty"`
	PopulateIndices bool   `json:"populate_indices,omitempty"`
	PopulateMatches bool   `json:"populate_matches,omitempty"`
}

type PatternAction struct {
	*PatternActionBase `json:",omitempty"`
}
