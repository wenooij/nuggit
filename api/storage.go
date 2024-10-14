package api

type StorageLite struct {
	*Ref `json:",omitempty"`
}

type StorageBase struct {
	LocalPath string `json:"local_path,omitempty"`
	InMemory  string `json:"in_memory,omitempty"`
}

type Storage struct {
	*StorageLite `json:",omitempty"`
	*StorageBase `json:",omitempty"`
}

type ImplicitStorageLite struct {
	Pipe         *PipeLite `json:"pipe,omitempty"`
	*StorageLite `json:",omitempty"`
}

type ImplicitStorage struct {
	*ImplicitStorageLite `json:",omitempty"`
	*StorageBase         `json:",omitempty"`
}

type StoreInterface interface {
	Resource(id string) *ResourceBase
	Node(id string) *NodeBase
	Pipe(id string) *PipeBase
	StoreResource(*Resource) error
	StoreNode(*Node) error
	StorePipe(*Pipe) error
}

type StorageAPI struct {
	store StoreInterface
}

func NewStorageAPI(store StoreInterface) *StorageAPI {
	return &StorageAPI{
		store: store,
	}
}
