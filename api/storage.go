package api

import "context"

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
	DeleteBatch(ctx context.Context, ids []string) error
}

type CollectionStore interface {
	StoreInterface[*Collection]
	ScanInterface[*CollectionLite]
	StoreBatchInterface[*CollectionLite]
}

type PipeStorage interface {
	StoreInterface[*PipeRich]
	ScanInterface[*PipeRich]
	ScanHostTriggered(ctx context.Context, hostname string, scanFn func(*PipeRich, error) error) error
}

type NodeStore interface {
	StoreInterface[*NodeRich]
	ScanInterface[*NodeRich]
}
