package api

import (
	"fmt"

	"github.com/wenooij/nuggit/status"
)

type Action struct {
	Action string         `json:"action,omitempty"`
	Spec   map[string]any `json:"spec,omitempty"`
}

func (a *Action) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *Action) GetSpec() map[string]any {
	if a == nil {
		return nil
	}
	return a.Spec
}

func validateAction(action *Action) error {
	method := action.GetAction()
	if method == "" {
		return fmt.Errorf("action must not be empty: %w", status.ErrInvalidArgument)
	}
	if _, ok := supportedActions[method]; !ok {
		return fmt.Errorf("action is not supported (%q): %w", method, status.ErrInvalidArgument)
	}
	return nil
}

const (
	ActionAttribute = "attribute" // AttributeAction extracts attribute names from HTML elements.
	ActionField     = "field"     // AttributeAction retrieves fields and or methods from HTML elements.
	ActionDocument  = "document"  // ActionDocument represents an action which copies the full document.
	ActionPattern   = "pattern"   // ActionPattern matches re2 patterns.
	ActionSelector  = "selector"  // ActionSelector matches CSS selectors.
	ActionPipe      = "pipe"      // ActionPipe executes the given pipeline in place.
	ActionExchange  = "exchange"  // ActionExchange submits the result to the server.
)

var supportedActions = map[string]struct{}{
	ActionAttribute: {},
	ActionField:     {},
	ActionDocument:  {},
	ActionPattern:   {},
	ActionSelector:  {},
	ActionPipe:      {},
	ActionExchange:  {},
}

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
}

type DocumentAction struct{}

type AttributeAction struct {
	Attribute string `json:"attribute,omitempty"`
}

type FieldAction struct {
	Field string `json:"field,omitempty"`
}

type PatternAction struct {
	Pattern         string `json:"pattern,omitempty"`
	Passthrough     bool   `json:"passthrough,omitempty"`
	PopulateIndices bool   `json:"populate_indices,omitempty"`
	PopulateMatches bool   `json:"populate_matches,omitempty"`
}

type PipeAction struct {
	Pipe string `json:"pipe,omitempty"`
}

type ExchangeAction struct {
	Pipe string `json:"pipe,omitempty"`
}
