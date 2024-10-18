package api

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit/status"
)

type Ref struct {
	ID   string `json:"id,omitempty"`
	URI  string `json:"uri,omitempty"`
	Name string `json:"name,omitempty"`
}

// precondition: uriBase should be formatted as "/api/API/".
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

func (r *Ref) GetID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

func (r *Ref) GetURI() string {
	if r == nil {
		return ""
	}
	return r.URI
}

func (r *Ref) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

type API struct {
	*CollectionsAPI
	*PipesAPI
	*TriggerAPI
}

func NewAPI(collectionStore CollectionStore, pipeStore PipeStore, triggerStore TriggerStore, resultStore StoreInterface[*TriggerResult]) *API {
	a := &API{
		CollectionsAPI: &CollectionsAPI{},
		PipesAPI:       &PipesAPI{},
		TriggerAPI:     &TriggerAPI{},
	}
	a.CollectionsAPI.Init(collectionStore)
	a.PipesAPI.Init(pipeStore)
	a.TriggerAPI.Init(triggerStore, resultStore, a.CollectionsAPI, a.PipesAPI)
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
