package api

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/wenooij/nuggit/status"
)

type Ref struct {
	NameDigest `json:","`
	ID         string `json:"id,omitempty"`
	URI        string `json:"uri,omitempty"`
}

func newRef(uriBase, id string) Ref {
	r := Ref{ID: id}
	_ = r.setURI(uriBase, id)
	return r
}

func newNamedRef(uriBase string, name NameDigest) Ref {
	r := Ref{NameDigest: name}
	_ = r.setURI(uriBase, name.String())
	return r
}

func (r *Ref) setURI(uriBase string, s string) error {
	if s == "" {
		return fmt.Errorf("identifier part must be set")
	}
	uri, err := url.JoinPath(uriBase, s)
	if err != nil {
		return err
	}
	r.URI = uri
	return nil
}

func (r *Ref) GetNameDigest() NameDigest {
	if r == nil {
		return NameDigest{}
	}
	return r.NameDigest
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

func compareRef(a, b Ref) int {
	if cmp := strings.Compare(a.ID, b.ID); cmp != 0 {
		return cmp
	}
	if cmp := compareNameDigest(a.GetNameDigest(), b.GetNameDigest()); cmp != 0 {
		return cmp
	}
	return strings.Compare(a.URI, b.URI)
}

type API struct {
	*CollectionsAPI
	*PipesAPI
	*TriggerAPI
}

type TriggerPlanner interface {
	Add(c *Collection, pipes []*Pipe) error
	Build() *TriggerPlan
}

func NewAPI(collectionStore CollectionStore, pipeStore PipeStore, triggerStore TriggerStore, newTriggerPlanner func() TriggerPlanner, resultStore StoreInterface[*TriggerResult]) *API {
	a := &API{
		CollectionsAPI: &CollectionsAPI{},
		PipesAPI:       &PipesAPI{},
		TriggerAPI:     &TriggerAPI{},
	}
	a.CollectionsAPI.Init(collectionStore)
	a.PipesAPI.Init(pipeStore)
	a.TriggerAPI.Init(triggerStore, newTriggerPlanner, resultStore, a.CollectionsAPI, a.PipesAPI)
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
