package api

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"time"

	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/trigger"
)

const triggersBaseURI = "/api/triggers"

func ValidateRule(c *trigger.Rule) error {
	if c == nil {
		return nil
	}
	if c.GetURLPattern() != "" {
		if c.GetHostname() == "" {
			return fmt.Errorf("url pattern requires a hostname to be provided: %w", status.ErrInvalidArgument)
		}
		if _, err := regexp.Compile(c.URLPattern); err != nil {
			return fmt.Errorf("url pattern is not a valid re2 (%q): %v: %w", c.URLPattern, err, status.ErrInvalidArgument)
		}
	}
	return nil
}

type TriggerEvent struct {
	Plan      string    `json:"plan,omitempty"`
	Implicit  bool      `json:"implicit,omitempty"`
	URL       string    `json:"url,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (e *TriggerEvent) GetPlan() string {
	if e == nil {
		return ""
	}
	return e.Plan
}

func (e *TriggerEvent) GetImplicit() bool {
	if e == nil {
		return false
	}
	return e.Implicit
}

func (e *TriggerEvent) GetURL() string {
	if e == nil {
		return ""
	}
	return e.URL
}

func (e *TriggerEvent) GetTimestamp() time.Time {
	if e == nil {
		return time.Time{}
	}
	return e.Timestamp
}

type TriggerResult struct {
	Pipe   integrity.NameDigest `json:"pipe,omitempty"`
	Result json.RawMessage      `json:"result,omitempty"`
}

func (r *TriggerResult) GetPipe() integrity.NameDigest {
	if r == nil {
		return integrity.NameDigest{}
	}
	return r.Pipe
}

func (r *TriggerResult) GetResult() json.RawMessage {
	if r == nil {
		return nil
	}
	return r.Result
}

type TriggerAPI struct {
	rule       RuleStore
	pipes      PipeStore
	plans      PlanStore
	results    ResultStore
	newPlanner func() TriggerPlanner
}

func (a *TriggerAPI) Init(rule RuleStore, pipes PipeStore, planStore PlanStore, resultStore ResultStore, newPlanner func() TriggerPlanner) {
	*a = TriggerAPI{
		rule:       rule,
		pipes:      pipes,
		plans:      planStore,
		results:    resultStore,
		newPlanner: newPlanner,
	}
}

type OpenTriggerRequest struct {
	URL          string                 `json:"url,omitempty"`
	Implicit     bool                   `json:"implicit,omitempty"`
	IncludePipes []integrity.NameDigest `json:"include_views,omitempty"`
	ExcludePipes []integrity.NameDigest `json:"exclude_views,omitempty"`
}

type OpenTriggerResponse struct {
	Trigger *Ref          `json:"trigger,omitempty"`
	Plan    *trigger.Plan `json:"plan,omitempty"`
}

func (a *TriggerAPI) OpenTrigger(ctx context.Context, req *OpenTriggerRequest) (*OpenTriggerResponse, error) {
	if err := provided("url", "is", req.URL); err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}

	pipes := make(map[integrity.NameDigest]*Pipe, 64)

	for pipe, err := range a.rule.ScanMatched(ctx, u) {
		if err != nil {
			return nil, err
		}
		pipes[pipe.GetNameDigest()] = pipe
	}

	if len(pipes) == 0 {
		// Plan is a no-op.
		// Don't store the trigger and send empty response.
		return &OpenTriggerResponse{}, nil
	}

	tp := a.newPlanner()

	// Add referenced pipes to Plan.
	// This is required for the FlattenPipes calls later on.
	for _, p := range pipes {
		// TODO: Maybe this query can be made batch?
		for rp, err := range a.pipes.ScanDependencies(ctx, p.GetNameDigest()) {
			if err != nil {
				return nil, err
			}
			tp.AddReferencedPipe(rp.NameDigest, rp.Pipe)
		}
	}

	// Add unique pipes to Plan.
	for p := range maps.Values(pipes) {
		if err := tp.AddPipe(p.NameDigest, p.Pipe); err != nil {
			return nil, err
		}
	}

	plan := tp.Build()
	if len(plan.GetSteps()) == 0 {
		// Plan is a no-op.
		// Don't store the trigger and send empty response.
		return &OpenTriggerResponse{}, nil
	}

	// Store the plan and return it since is isn't a no-op.
	planRef, err := newRef(triggersBaseURI)
	if err != nil {
		return nil, err
	}

	if err := a.plans.Store(ctx, planRef.ID, plan); err != nil {
		return nil, err
	}

	return &OpenTriggerResponse{
		Trigger: &planRef,
		Plan:    plan,
	}, nil
}

type ExchangeResultsRequest struct {
	Trigger *TriggerEvent   `json:"trigger,omitempty"`
	Results []TriggerResult `json:"results,omitempty"`
}

type ExchangeResultsResponse struct{}

func (a *TriggerAPI) ExchangeResults(ctx context.Context, req *ExchangeResultsRequest) (*ExchangeResultsResponse, error) {
	if err := provided("trigger", "is", req.Trigger); err != nil {
		return nil, err
	}
	if err := a.results.StoreResults(ctx, req.Trigger, req.Results); err != nil {
		return nil, err
	}
	return &ExchangeResultsResponse{}, nil
}

type CloseTriggerRequest struct {
	Trigger string `json:"trigger,omitempty"`
}

type CloseTriggerResponse struct{}

func (a *TriggerAPI) CloseTrigger(ctx context.Context, req *CloseTriggerRequest) (*CloseTriggerResponse, error) {
	if err := a.plans.Finish(ctx, req.Trigger); err != nil {
		return nil, err
	}
	return &CloseTriggerResponse{}, nil
}

type CreateRuleRequest struct {
	Pipe *integrity.NameDigest `json:"pipe,omitempty"`
	Rule *trigger.Rule         `json:"rule,omitempty"`
}

type CreateRuleResponse struct{}

func (a *TriggerAPI) CreateRule(ctx context.Context, req *CreateRuleRequest) (*CreateRuleResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	if err := provided("digest", "is", req.Pipe.Digest); err != nil {
		return nil, err
	}
	if err := provided("rule", "is", req.Rule); err != nil {
		return nil, err
	}
	if err := ValidateRule(req.Rule); err != nil {
		return nil, err
	}
	if err := a.rule.StoreRule(ctx, *req.Pipe, req.Rule); err != nil {
		return nil, err
	}
	return &CreateRuleResponse{}, nil
}
