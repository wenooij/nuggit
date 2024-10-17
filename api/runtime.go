package api

import (
	"context"
)

type RuntimeLite struct {
	*Ref `json:",omitempty"`
}

func NewRuntimeLite(id string) *RuntimeLite {
	return &RuntimeLite{newRef("/api/runtimes/", id)}
}

type RuntimeBase struct {
	Name             string        `json:"name,omitempty"`
	SupportedActions []*ActionLite `json:"supported_actions,omitempty"`
}

type Runtime struct {
	*RuntimeLite `json:",omitempty"`
	*RuntimeBase `json:",omitempty"`
}

type RuntimesAPI struct{}

type RuntimeStatusRequest struct{}

type RuntimeStatusResponse struct{}

func (a *RuntimesAPI) RuntimeStatus(context.Context, *RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	return &RuntimeStatusResponse{}, nil
}
