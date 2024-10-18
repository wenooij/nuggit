package api

import (
	"bytes"
	"encoding/json"
)

type Action struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

func (a *Action) sameAction(b *Action) bool {
	return a == nil && b == nil ||
		a != nil && b != nil && a.Action == b.Action && bytes.Equal(a.Spec, b.Spec)
}

func (a *Action) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *Action) GetSpec() json.RawMessage {
	if a == nil {
		return nil
	}
	return a.Spec
}

const (
	ActionAttribute = "attribute" // AttributeAction extracts attribute names from the
	ActionDocument  = "document"  // ActionDocument represents an action which copies the full document.
	ActionPattern   = "pattern"   // ActionPattern matches re2 patterns.
	ActionSelector  = "selector"  // ActionSelector matches CSS selectors.
	ActionExchange  = "exchange"  // ActionExchange submits the result to the server.
)

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
}

type DocumentAction struct{}

type AttributeAction struct {
	Attribute string `json:"attribute,omitempty"`
}

type PatternAction struct {
	Pattern         string `json:"pattern,omitempty"`
	Passthrough     bool   `json:"passthrough,omitempty"`
	PopulateIndices bool   `json:"populate_indices,omitempty"`
	PopulateMatches bool   `json:"populate_matches,omitempty"`
}

type ExchangeAction struct {
	Pipe string `json:"pipe,omitempty"`
}
