package api

import (
	"context"
	"iter"
	"net/url"

	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/trigger"
)

type PipeStore interface {
	Load(ctx context.Context, pipe integrity.NameDigest) (*Pipe, error)
	Store(context.Context, *Pipe) error
	StoreBatch(context.Context, []*Pipe) error
	ScanNames(context.Context) iter.Seq2[integrity.NameDigest, error]
	Scan(context.Context) iter.Seq2[*Pipe, error]
	ScanDependencies(ctx context.Context, pipe integrity.NameDigest) iter.Seq2[*Pipe, error]
	Disable(ctx context.Context, pipe integrity.NameDigest) error
	Enable(ctx context.Context, pipe integrity.NameDigest) error
}

type RuleStore interface {
	StoreRule(ctx context.Context, pipe integrity.NameDigest, rule *trigger.Rule) error
	DeleteRule(ctx context.Context, pipe integrity.NameDigest, rule *trigger.Rule) error
	ScanMatched(ctx context.Context, u *url.URL) iter.Seq2[*Pipe, error]
}

type PlanStore interface {
	Store(ctx context.Context, uuid string, plan *trigger.Plan) error
	Finish(ctx context.Context, uuid string) error
}

type ResultStore interface {
	StoreResults(ctx context.Context, trigger *TriggerEvent, results []TriggerResult) error
}

type ResourceStore interface {
	StorePipeResource(context.Context, *Resource, *Pipe) error
	StoreViewResource(ctx context.Context, r *Resource, viewUUID string) error
}

type ViewStore interface {
	Store(ctx context.Context, uuid string, view *View) error
}
