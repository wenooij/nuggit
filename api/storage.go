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

type UUID interface {
	UUID() string
}

type StoreInterface[T UUID] interface {
	Len(ctx context.Context) (n int, exact bool)
	Exists(ctx context.Context, id string) (bool, error)
	Load(ctx context.Context, id string) (T, error)
	Delete(ctx context.Context, id string) error
	Store(ctx context.Context, t T) error
	StoreOrReplace(ctx context.Context, t T) error
}

type ScanInterface[T UUID] interface {
	Scan(ctx context.Context, scanFn func(T, error) error) error
}

type StoreBatchInterface[T UUID] interface {
	LoadBatch(ctx context.Context, ids []string) ([]T, error)
	DeleteBatch(ctx context.Context, ids []string) error
}

type CollectionStore interface {
	StoreInterface[*Collection]
	StoreBatchInterface[*Collection]
	ScanInterface[*CollectionLite]
	ScanTriggered(ctx context.Context, u *url.URL, scanFn func(*Collection, error) error) error
}

type PipeStorage interface {
	StoreInterface[*Pipe]
	StoreBatchInterface[*Pipe]
	ScanInterface[*Pipe]
}
