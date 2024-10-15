package api

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit/status"
)

type Ref struct {
	ID  string `json:"id,omitempty"`
	URI string `json:"uri,omitempty"`
}

func (r *Ref) UUID() string { return r.ID }

type API struct {
	*ActionsAPI
	// *ClientAPI
	*NodesAPI
	*PipesAPI
	*ResourcesAPI
	*RuntimesAPI
	*TriggerAPI
	mu sync.Mutex // For methods reading and writing across API boundaries.
}

func NewAPI(storeType StorageType) (*API, error) {
	a := &API{
		ActionsAPI:   &ActionsAPI{},
		NodesAPI:     &NodesAPI{},
		PipesAPI:     &PipesAPI{},
		ResourcesAPI: &ResourcesAPI{},
		RuntimesAPI:  &RuntimesAPI{},
		TriggerAPI:   &TriggerAPI{},
	}
	a.NodesAPI.Init(a, a.PipesAPI, storeType)
	if err := a.PipesAPI.Init(a, a.NodesAPI, storeType); err != nil {
		return nil, err
	}
	a.ResourcesAPI.Init(storeType)
	if err := a.RuntimesAPI.Init(storeType); err != nil {
		return nil, err
	}
	a.TriggerAPI.Init(a, a.RuntimesAPI, a.PipesAPI)
	return a, nil
}

func exclude(arg string, are string, t any) error {
	if v := reflect.ValueOf(t); !v.IsZero() {
		return fmt.Errorf("%s %s not allowed here: %w", arg, are, status.ErrInvalidArgument)
	}
	return nil
}

func provided(arg string, is string, t any) error {
	if v := reflect.ValueOf(t); v.IsZero() {
		return fmt.Errorf("%s %s required: %w", arg, is, status.ErrInvalidArgument)
	}
	return nil
}

func newUUID(uniqueCheck func(id string) bool) (string, error) {
	const maxAttempts = 100
	for attempts := maxAttempts; attempts > 0; attempts-- {
		u, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("%v: %w", err, status.ErrInternal)
		}
		if id := u.String(); uniqueCheck(id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to generate a unique ID after %d attempts: %w", maxAttempts, status.ErrInternal)
}

func validateUUID(s string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	return nil
}
