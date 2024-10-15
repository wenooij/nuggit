package api

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/wenooij/nuggit/status"
)

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
	ActionDocument = "document" // ActionDocument represents an action which copies the full document.
	ActionExchange = "exchange" // ActionExchange marks the (network) boundary for a client-server data exchange.
	ActionPattern  = "pattern"  // ActionPattern matches re2 patterns.
	ActionSelector = "selector" // ActionSelector matches CSS selectors.
	ActionExport   = "export"   // ActionExport exports points to a collection.
)

func builtinActions() []string {
	return []string{
		ActionDocument,
		ActionExchange,
		ActionPattern,
		ActionSelector,
		ActionExport,
	}
}

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
	Raw      bool   `json:"raw,omitempty"`
}

type DocumentAction struct {
	Raw bool `json:"raw,omitempty"`
}

type ExchangeAction struct {
	Next string `json:"next,omitempty"`
	Args *Args  `json:"args,omitempty"`
}

type ExportAction struct {
	Nullable        bool   `json:"nullable,omitempty"`
	IncludeMetadata bool   `json:"include_metadata,omitempty"`
	Point           *Point `json:"point,omitempty"`
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

func (a *PatternAction) Run(args *Args) (*BatchArgs, error) {
	if a == nil || args == nil {
		return &BatchArgs{}, nil
	}
	s := args.String
	res := &BatchArgs{Args: &Args{Indices: make([]int, 0, 64)}}
	match := a.expr.FindAllStringSubmatchIndex(s, -1)
	for _, m := range match {
		var ms []int
		if a.submatch {
			ms = m[2:]
		} else {
			ms = m[:2]
		}
		if a.PopulateIndices {
			res.Indices = append(res.Indices, ms...)
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

func (*ActionsAPI) ListBuiltinActions(*ListBuiltinActionsRequest) (*ListBuiltinActionsResponse, error) {
	return &ListBuiltinActionsResponse{Actions: builtinActions()}, nil
}

type RunActionRequest struct {
	*Action `json:",omitempty"`
	Args    *Args `json:"args,omitempty"`
}

type RunActionResponse struct {
	Results *BatchArgs `json:"results,omitempty"`
}

func (*ActionsAPI) RunAction(req *RunActionRequest) (*RunActionResponse, error) {
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
