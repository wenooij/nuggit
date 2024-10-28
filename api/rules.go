package api

import (
	"context"

	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/trigger"
)

const rulesBaseURI = "/api/rules"

type RulesAPI struct {
	rules RuleStore
}

func (a *RulesAPI) Init(rules RuleStore) {
	*a = RulesAPI{
		rules: rules,
	}
}

type CreateRuleRequest struct {
	Pipe string        `json:"pipe,omitempty"`
	Rule *trigger.Rule `json:"rule,omitempty"`
}

type CreateRuleResponse struct{}

func (a *RulesAPI) CreateRule(ctx context.Context, req *CreateRuleRequest) (*CreateRuleResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	nameDigest, err := integrity.ParseNameDigest(req.Pipe)
	if err != nil {
		return nil, err
	}
	if err := provided("rule", "is", req.Rule); err != nil {
		return nil, err
	}
	if err := ValidateRule(req.Rule); err != nil {
		return nil, err
	}
	if err := a.rules.StoreRule(ctx, nameDigest, req.Rule); err != nil {
		return nil, err
	}
	return &CreateRuleResponse{}, nil
}

type DeleteRuleRequest struct {
	Pipe string        `json:"pipe,omitempty"`
	Rule *trigger.Rule `json:"rule,omitempty"`
}

type DeleteRuleResponse struct{}

func (a *RulesAPI) DeleteRule(ctx context.Context, req *DeleteRuleRequest) (*DeleteRuleResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	nameDigest, err := integrity.ParseNameDigest(req.Pipe)
	if err != nil {
		return nil, err
	}
	if err := provided("rule", "is", req.Rule); err != nil {
		return nil, err
	}
	if err := ValidateRule(req.Rule); err != nil {
		return nil, err
	}
	if err := a.rules.DeleteRule(ctx, nameDigest, req.Rule); err != nil {
		return nil, err
	}
	return &DeleteRuleResponse{}, nil
}
