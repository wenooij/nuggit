package api

import "encoding/json"

type ActionLite struct {
	*Ref `json:",omitempty"`
}

type ActionBase struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

type Action struct {
	*ActionLite `json:",omitempty"`
	*ActionBase `json:",omitempty"`
}

const (
	ActionUndefined   = ""            // Same as ActionPassthrough.
	ActionPassthrough = "passthrough" // ActionPassthrough does nothing.
	ActionDocument    = "document"    // ActionDocument represents an action which copies the full document.
	ActionExchange    = "exchange"    // ActionExchange marks the (network) boundary for a client-server data exchange.
	ActionLiteral     = "literal"     // ActionLiteral matches string literals.
	ActionPattern     = "pattern"     // ActionPattern matches re2 patterns.
	ActionSelector    = "selector"    // ActionSelector matches CSS selectors.
)

func builtinActions() []string {
	return []string{
		ActionPassthrough,
		ActionDocument,
		ActionExchange,
		ActionLiteral,
		ActionPattern,
		ActionSelector,
	}
}

type ActionsAPI struct{}

type ListBuiltinActionsRequest struct{}

type ListBuiltinActionsResponse struct {
	Actions []string `json:"actions,omitempty"`
}

func (*ActionsAPI) ListBuiltinActions(*ListBuiltinActionsRequest) (*ListBuiltinActionsResponse, error) {
	return &ListBuiltinActionsResponse{Actions: builtinActions()}, nil
}
