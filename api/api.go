package api

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit/status"
)

type Ref struct {
	ID  string `json:"id,omitempty"`
	URI string `json:"uri,omitempty"`
}

func newRef(uriBase, id string) *Ref {
	if id == "" {
		return &Ref{}
	}
	if _, err := uuid.Parse(id); err != nil {
		return &Ref{ID: id}
	}
	return &Ref{
		ID:  id,
		URI: fmt.Sprint(uriBase, id),
	}
}

func (r *Ref) UUID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

type API struct {
	*CollectionsAPI
	*NodesAPI
	*PipesAPI
	*ResourcesAPI
	*RuntimesAPI
	*TriggerAPI
}

func NewAPI(collectionStore CollectionStore, pipeStore PipeStorage, nodeStore NodeStore) (*API, error) {
	a := &API{
		CollectionsAPI: &CollectionsAPI{},
		NodesAPI:       &NodesAPI{},
		PipesAPI:       &PipesAPI{},
		ResourcesAPI:   &ResourcesAPI{},
		RuntimesAPI:    &RuntimesAPI{},
		TriggerAPI:     &TriggerAPI{},
	}
	if err := a.CollectionsAPI.Init(collectionStore); err != nil {
		return nil, err
	}
	if err := a.NodesAPI.Init(nodeStore, a.PipesAPI); err != nil {
		return nil, err
	}
	a.PipesAPI.Init(pipeStore, a.NodesAPI)
	a.TriggerAPI.Init(a.RuntimesAPI, a.PipesAPI)
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

func newUUID(uniqueCheck func(id string) error) (string, error) {
	const maxAttempts = 3
	var lastErr error
	for attempts := maxAttempts; attempts > 0; attempts-- {
		u, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("%v: %w", err, status.ErrInternal)
		}
		id := u.String()
		if err := uniqueCheck(id); errors.Is(err, status.ErrNotFound) {
			return id, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = status.ErrAlreadyExists
	}
	return "", fmt.Errorf("failed to generate a unique ID after %d attempts: %w", maxAttempts, lastErr)
}

func validateUUID(s string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	return nil
}
