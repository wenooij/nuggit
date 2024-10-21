package api

import (
	"context"
	"iter"
	"net/url"
)

type StorageOpStatus = string

const (
	StorageOpUndefined StorageOpStatus = "" // Same as StatusUnknown.
	StorageOpUnknown   StorageOpStatus = "unknown"
	StorageOpStoring   StorageOpStatus = "storing"
	StorageOpComplete  StorageOpStatus = "complete"
)

type StoreInterface[T any] interface {
	Load(ctx context.Context, id string) (T, error)
	Delete(ctx context.Context, id string) error
	Store(ctx context.Context, t T) (string, error)
}

type ScanInterface[T any] interface {
	Scan(ctx context.Context, scanFn func(T, error) error) error
}

type StoreNamed[T interface{ GetName() string }] interface {
	Load(ctx context.Context, name NameDigest) (T, error)
	Store(ctx context.Context, t T) (NameDigest, error)
	Delete(ctx context.Context, name NameDigest) error
	ScanNames(ctx context.Context) iter.Seq2[NameDigest, error]
	Scan(ctx context.Context) iter.Seq2[struct {
		NameDigest
		Elem T
	}, error]
}

type StoreNamedBatch[T interface{ GetName() string }] interface {
	LoadBatch(ctx context.Context, names []NameDigest) iter.Seq2[T, error]
	StoreBatch(ctx context.Context, t []T) ([]NameDigest, error)
	DeleteBatch(ctx context.Context, nd []NameDigest) error
}

type ScanNamed[T interface{ GetName() string }] interface {
	ScanNames(ctx context.Context) iter.Seq2[NameDigest, error]
	Scan(ctx context.Context) iter.Seq2[struct {
		NameDigest
		Elem T
	}, error]
}

type CollectionStore interface {
	StoreNamed[*Collection]
	StoreNamedBatch[*Collection]
	ScanNamed[*Collection]
	ScanTriggered(ctx context.Context, u *url.URL) iter.Seq2[struct {
		*Collection
		*Pipe
	}, error]
}

type PipeStore interface {
	StoreNamed[*Pipe]
	StoreNamedBatch[*Pipe]
	ScanNamed[*Pipe]
}

type TriggerStore interface {
	StoreInterface[*TriggerRecord]
}
