package api

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"time"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/trigger"
)

const triggersBaseURI = "/api/triggers"

func ValidateRule(c nuggit.Rule) error {
	if c.GetURLPattern() != "" {
		if c.GetHostname() == "" {
			return fmt.Errorf("url pattern requires a hostname to be provided: %w", status.ErrInvalidArgument)
		}
		if _, err := regexp.Compile(c.URLPattern); err != nil {
			return fmt.Errorf("url pattern is not a valid re2 (%q): %v: %w", c.URLPattern, err, status.ErrInvalidArgument)
		}
	}
	// TODO: Validate hostname.
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

// TODO: Add Point to this struct.
type TriggerResult struct {
	Pipe   string          `json:"pipe,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func (r *TriggerResult) GetPipe() string {
	if r == nil {
		return ""
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
	rules      RuleStore
	pipes      PipeStore
	plans      PlanStore
	results    ResultStore
	newPlanner func() TriggerPlanner
}

func (a *TriggerAPI) Init(rules RuleStore, pipes PipeStore, planStore PlanStore, resultStore ResultStore, newPlanner func() TriggerPlanner) {
	*a = TriggerAPI{
		rules:      rules,
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

	for pipe, err := range a.rules.ScanMatched(ctx, u) {
		if err != nil {
			return nil, err
		}
		pipes[integrity.Key(pipe)] = pipe
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
		for rp, err := range a.pipes.ScanDependencies(ctx, integrity.Key(p)) {
			if err != nil {
				return nil, err
			}
			tp.AddReferencedPipe(rp.GetName(), rp.GetDigest(), rp.Pipe)
		}
	}

	// Add unique pipes to Plan.
	for p := range maps.Values(pipes) {
		if err := tp.AddPipe(p.GetName(), p.GetDigest(), p.Pipe); err != nil {
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
