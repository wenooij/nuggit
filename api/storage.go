package api

type StorageType = string

const (
	StorageUndefined StorageType = "" // Same as in memory.
	StorageInMemory  StorageType = "inmemory"
)

type StorageOpStatus = string

const (
	StorageOpUndefined StorageOpStatus = "" // Same as StatusUnknown.
	StorageOpUnknown   StorageOpStatus = "unknown"
	StorageOpStoring   StorageOpStatus = "storing"
	StorageOpComplete  StorageOpStatus = "complete"
)

type StorageOpLite struct {
	*Ref `json:",omitempty"`
}

type UUID interface {
	UUID() string
}

type StoreInterface[T UUID] interface {
	Len() (n int, exact bool)
	Load(id string) (T, error)
	Scan(func(T, error) error) error
	Delete(id string) (*StorageOpLite, error)
	Store(T) (*StorageOpLite, error)
	StoreOrReplace(T) (*StorageOpLite, error)
	Poll(storageOpID string) (StorageOpStatus, error)
}
