package api

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/wenooij/nuggit/status"
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

func builtinActions() []string {
	return []string{
		ActionDocument,
		ActionExchange,
		ActionPattern,
		ActionSelector,
	}
}

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
	Raw      bool   `json:"raw,omitempty"`
}

type DocumentAction struct {
	Raw bool `json:"raw,omitempty"`
}

type AttributeAction struct {
	Attribute string `json:"attribute,omitempty"`
}

type ExchangeAction struct {
	Next string `json:"next,omitempty"`
	Args *Args  `json:"args,omitempty"`
}

type PatternActionBase struct {
	Pattern         string `json:"pattern,omitempty"`
	Passthrough     bool   `json:"passthrough,omitempty"`
	PopulateIndices bool   `json:"populate_indices,omitempty"`
	PopulateMatches bool   `json:"populate_matches,omitempty"`
}

type PatternActionInternal struct {
	expr     *regexp.Regexp
	submatch bool
}

type PatternAction struct {
	*PatternActionBase     `json:",omitempty"`
	*PatternActionInternal `json:"-"`
}

func newPatternAction(a *PatternActionBase) (*PatternAction, error) {
	expr, err := regexp.Compile(a.Pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pattern: %v: %w", err, status.ErrInvalidArgument)
	}
	if expr.NumSubexp() > 1 {
		return nil, fmt.Errorf("pattern may only use 1 capturing group: %w", status.ErrInvalidArgument)
	}
	return &PatternAction{
		PatternActionBase:     a,
		PatternActionInternal: &PatternActionInternal{expr: expr, submatch: expr.NumSubexp() == 1},
	}, nil
}

func (a *PatternAction) Run(args *BatchArgs) (*BatchArgs, error) {
	if a == nil || args == nil {
		return &BatchArgs{}, nil
	}
	s := args.String
	res := &BatchArgs{Int64s: make([]int64, 0, 64)}
	match := a.expr.FindAllStringSubmatchIndex(s, -1)
	for _, m := range match {
		var ms []int
		if a.submatch {
			ms = m[2:]
		} else {
			ms = m[:2]
		}
		if a.PopulateIndices {
			res.Int64s = append(res.Int64s, int64(ms[0]), int64(ms[1]))
		}
		if a.PopulateMatches {
			res.Strings = append(res.Strings, s[ms[0]:ms[1]])
		}
	}
	if a.Passthrough {
		res.Args.String = s
	}
	return res, nil
}

type ActionsAPI struct{}

type ListBuiltinActionsRequest struct{}

type ListBuiltinActionsResponse struct {
	Actions []string `json:"actions,omitempty"`
}

func (*ActionsAPI) ListBuiltinActions(context.Context, *ListBuiltinActionsRequest) (*ListBuiltinActionsResponse, error) {
	return &ListBuiltinActionsResponse{Actions: builtinActions()}, nil
}

type RunActionRequest struct {
	*Action `json:",omitempty"`
	Args    *BatchArgs `json:"args,omitempty"`
}

type RunActionResponse struct {
	Results     *BatchArgs   `json:"results,omitempty"`
	PointValues *PointValues `json:"point_values,omitempty"`
}

func (*ActionsAPI) RunAction(ctx context.Context, req *RunActionRequest) (*RunActionResponse, error) {
	if err := provided("action", "is", req.Action); err != nil {
		return nil, err
	}
	if err := provided("action", "is", req.Action.ActionBase); err != nil {
		return nil, err
	}
	if req.Action.ActionLite != nil {
		return nil, fmt.Errorf("custom actions are not yet supported: %w", status.ErrUnimplemented)
	}
	unmarshalSpec := func(v any) error {
		if err := json.Unmarshal(req.Action.Spec, v); err != nil {
			return fmt.Errorf("failed to unmarshal action (did you forget to provide a spec?): %v: %w", err, status.ErrInvalidArgument)
		}
		return nil
	}
	switch req.Action.Action {
	case ActionDocument, ActionExchange, ActionSelector:
		return nil, fmt.Errorf("client action cannot be tested: %w", status.ErrInvalidArgument)
	case ActionPattern:
		spec := new(PatternActionBase)
		if err := unmarshalSpec(spec); err != nil {
			return nil, err
		}
		p, err := newPatternAction(spec)
		if err != nil {
			return nil, err
		}
		res, err := p.Run(req.Args)
		if err != nil {
			return nil, err
		}
		return &RunActionResponse{Results: res}, nil
	default:
		return nil, fmt.Errorf("unknown action: %w", status.ErrInvalidArgument)
	}
}
