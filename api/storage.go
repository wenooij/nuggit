package api

import (
	"context"
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

type ScanRefInterface interface {
	ScanRef(ctx context.Context, scanFn func(string, error) error) error
}

type StoreBatchInterface[T any] interface {
	LoadBatch(ctx context.Context, ids []string) ([]T, []string, error)
	DeleteBatch(ctx context.Context, ids []string) error
}

type Lookup[T any] interface {
	Lookup(ctx context.Context, name string) (string, error)
}

type CollectionStore interface {
	StoreInterface[*Collection]
	StoreBatchInterface[*Collection]
	Lookup[*Collection]
	ScanRefInterface
	ScanTriggered(ctx context.Context, u *url.URL, scanFn func(id string, collection *Collection, pipes []*Pipe, err error) error) error
}

type PipeStore interface {
	StoreInterface[*Pipe]
	StoreBatchInterface[*Pipe]
	ScanInterface[*Pipe]
	ScanRefInterface
}

type TriggerStore interface {
	StoreInterface[*TriggerRecord]
}
