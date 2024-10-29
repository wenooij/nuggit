package api

import (
	"context"

	"github.com/wenooij/nuggit"
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
	Rule *nuggit.Rule `json:"rule,omitempty"`
}

type CreateRuleResponse struct{}

func (a *RulesAPI) CreateRule(ctx context.Context, req *CreateRuleRequest) (*CreateRuleResponse, error) {
	if err := provided("rule", "is", req.Rule); err != nil {
		return nil, err
	}
	if err := ValidateRule(*req.Rule); err != nil {
		return nil, err
	}
	if err := a.rules.StoreRule(ctx, *req.Rule); err != nil {
		return nil, err
	}
	return &CreateRuleResponse{}, nil
}

type DeleteRuleRequest struct {
	Rule *nuggit.Rule `json:"rule,omitempty"`
}

type DeleteRuleResponse struct{}

func (a *RulesAPI) DeleteRule(ctx context.Context, req *DeleteRuleRequest) (*DeleteRuleResponse, error) {
	if err := provided("rule", "is", req.Rule); err != nil {
		return nil, err
	}
	if err := ValidateRule(*req.Rule); err != nil {
		return nil, err
	}
	if err := a.rules.DeleteRule(ctx, *req.Rule); err != nil {
		return nil, err
	}
	return &DeleteRuleResponse{}, nil
}
