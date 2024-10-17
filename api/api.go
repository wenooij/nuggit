package api

import (
	"context"
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
	*PipesAPI
	*TriggerAPI
}

func NewAPI(collectionStore CollectionStore, pipeStore PipeStorage, triggerStore StoreInterface[*Trigger]) *API {
	a := &API{
		CollectionsAPI: &CollectionsAPI{},
		PipesAPI:       &PipesAPI{},
		TriggerAPI:     &TriggerAPI{},
	}
	a.CollectionsAPI.Init(collectionStore)
	a.PipesAPI.Init(pipeStore)
	a.TriggerAPI.Init(triggerStore, a.CollectionsAPI, a.PipesAPI)
	return a
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

func newUUID(ctx context.Context, existsCheck func(ctx context.Context, id string) (bool, error)) (string, error) {
	const maxAttempts = 3
	for attempts := maxAttempts; attempts > 0; attempts-- {
		u, err := uuid.NewV7()
		if err != nil {
			return "", fmt.Errorf("%v: %w", err, status.ErrInternal)
		}
		id := u.String()
		exists, err := existsCheck(ctx, id)
		if err != nil {
			return "", err
		}
		if exists {
			continue
		}
		return id, nil
	}
	return "", fmt.Errorf("failed to generate a unique ID after %d attempts", maxAttempts)
}

func validateUUID(s string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	return nil
}
