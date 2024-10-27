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
	integrity.NameDigest `json:","`
	ID                   string `json:"id,omitempty"`
	URI                  string `json:"uri,omitempty"`
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

func newNamedRef(uriBase string, name integrity.NameDigest) Ref {
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

func (r *Ref) GetNameDigest() integrity.NameDigest {
	if r == nil {
		return integrity.NameDigest{}
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
	if cmp := integrity.CompareNameDigest(a.GetNameDigest(), b.GetNameDigest()); cmp != 0 {
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
	AddReferencedPipe(integrity.NameDigest, nuggit.Pipe)
	AddPipe(integrity.NameDigest, nuggit.Pipe) error
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
