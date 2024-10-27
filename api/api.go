package api

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/trigger"
)

type Ref struct {
	Name   string `json:"name,omitempty"`
	Digest string `json:"digest,omitempty"`
	ID     string `json:"id,omitempty"`
	URI    string `json:"uri,omitempty"`
}

func newRef(uriBase string) (Ref, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return Ref{}, fmt.Errorf("%v: %w", err, status.ErrInternal)
	}
	return newRefID(uriBase, u.String()), nil
}

func newRefID(uriBase, id string) Ref {
	r := Ref{ID: id}
	_ = r.setURI(uriBase, id)
	return r
}

func newNamedRef(uriBase string, nd integrity.NameDigest) Ref {
	var r Ref
	if name, err := integrity.FormatString(nd); err == nil {
		r.Name = nd.GetName()
		r.Digest = nd.GetDigest()
		_ = r.setURI(uriBase, name)
	}
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

func (r *Ref) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *Ref) GetDigest() string {
	if r == nil {
		return ""
	}
	return r.Digest
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
	if cmp := strings.Compare(a.Name, b.Name); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(a.Digest, b.Digest); cmp != 0 {
		return cmp
	}
	return strings.Compare(a.URI, b.URI)
}

type API struct {
	*ViewsAPI
	*PipesAPI
	*TriggerAPI
}

type TriggerPlanner interface {
	AddReferencedPipe(name, digest string, pipe nuggit.Pipe)
	AddPipe(name, digest string, pipe nuggit.Pipe) error
	Build() *trigger.Plan
}

func NewAPI(viewStore ViewStore, pipeStore PipeStore, ruleStore RuleStore, planStore PlanStore, resultStore ResultStore, newTriggerPlanner func() TriggerPlanner) *API {
	a := &API{
		ViewsAPI:   &ViewsAPI{},
		PipesAPI:   &PipesAPI{},
		TriggerAPI: &TriggerAPI{},
	}
	a.ViewsAPI.Init(viewStore, pipeStore)
	a.PipesAPI.Init(pipeStore, ruleStore)
	a.TriggerAPI.Init(ruleStore, pipeStore, planStore, resultStore, newTriggerPlanner)
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
