package api

import (
	"errors"
	"fmt"
	"sync"

	"github.com/wenooij/nuggit/status"
)

type storageInMemory[T UUID] struct {
	objects map[string]T
	mu      sync.RWMutex
}

func newStorageInMemory[T UUID]() *storageInMemory[T] {
	return &storageInMemory[T]{
		objects: make(map[string]T),
	}
}

func (m *storageInMemory[T]) Load(id string) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	object, ok := m.objects[id]
	if !ok {
		var zero T
		return zero, fmt.Errorf("failed to load object: %w", status.ErrNotFound)
	}
	return object, nil
}

var ErrStopScan = errors.New("stop scan")

func (m *storageInMemory[T]) Len() (int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.objects), true
}

func (m *storageInMemory[T]) Scan(scanFn func(t T, err error) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.objects {
		if err := scanFn(t, nil); err != nil {
			if err == ErrStopScan {
				break
			}
			return err
		}
	}
	return nil
}

func (m *storageInMemory[T]) Poll(id string) (StorageOpStatus, error) {
	// Operation is always complete with in memory storage.
	return StorageOpComplete, nil
}

func (m *storageInMemory[T]) Delete(id string) (*StorageOpLite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, id)
	return nil, nil
}

func (m *storageInMemory[T]) Store(object T) (*StorageOpLite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oid := object.UUID()
	if oid == "" {
		return nil, fmt.Errorf("failed to store object: invalid empty ref: %w", status.ErrInvalidArgument)
	}
	if _, ok := m.objects[oid]; ok {
		return nil, fmt.Errorf("failed to store object: %w", status.ErrAlreadyExists)
	}
	m.objects[oid] = object
	return nil, nil
}

func (m *storageInMemory[T]) StoreOrReplace(object T) (*StorageOpLite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oid := object.UUID()
	if oid == "" {
		return nil, fmt.Errorf("failed to store object: invalid empty ref: %w", status.ErrInvalidArgument)
	}
	m.objects[oid] = object
	return nil, nil
}

type indexInMemory struct {
	indices map[string]map[string]struct{}
	mu      sync.RWMutex
}

func newIndexInMemory() *indexInMemory {
	return &indexInMemory{indices: make(map[string]map[string]struct{})}
}

func (m *indexInMemory) ScanKey(key string, scanFn func(string, error) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, index := range m.indices {
		for v := range index {
			if err := scanFn(v, nil); err != nil {
				if err == ErrStopScan {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (m *indexInMemory) DeleteKeyValue(key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	index := m.indices[key]
	delete(index, value)
	return nil
}
