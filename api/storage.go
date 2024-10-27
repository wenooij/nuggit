package api

import (
	"context"
	"iter"
	"net/url"

	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/trigger"
)

type PipeStore interface {
	Load(context.Context, integrity.NameDigest) (*Pipe, error)
	Store(context.Context, *Pipe) error
	StoreBatch(context.Context, []*Pipe) error
	ScanNames(context.Context) iter.Seq2[integrity.NameDigest, error]
	Scan(context.Context) iter.Seq2[*Pipe, error]
	ScanDependencies(context.Context, integrity.NameDigest) iter.Seq2[*Pipe, error]
}

type RuleStore interface {
	Disable(context.Context, integrity.NameDigest) error
	StoreRule(ctx context.Context, pipe integrity.NameDigest, rule *trigger.Rule) error
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
	Load(context.Context, integrity.NameDigest) (*Resource, error)
	Delete(context.Context, integrity.NameDigest) error
	Store(context.Context, *Resource) error
	Scan(context.Context) iter.Seq2[*Resource, error]
}

type ViewStore interface {
	Store(ctx context.Context, uuid string, view *View) error
}
