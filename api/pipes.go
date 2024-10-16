package api

import (
	"errors"
	"fmt"

	"github.com/wenooij/nuggit/status"
)

type PipeLite struct {
	*Ref `json:",omitempty"`
}

func newPipeLite(id string) *PipeLite {
	return &PipeLite{&Ref{
		ID:  id,
		URI: fmt.Sprintf("/api/pipes/%s", id),
	}}
}

type PipeBase struct {
	Sequence   []*NodeLite `json:"sequence,omitempty"`
	Conditions *Conditions `json:"conditions,omitempty"`
}

type Pipe struct {
	*PipeLite `json:",omitempty"`
	*PipeBase `json:",omitempty"`
}

type PipeState struct {
	Disabled    bool                `json:"disabled,omitempty"`
	Collections map[string]struct{} `json:"collections,omitempty"`
}

type PipeBaseRich struct {
	Sequence   []*NodeBase `json:"sequence,omitempty"`
	Conditions *Conditions `json:"conditions,omitempty"`
}

type PipeRich struct {
	*PipeLite  `json:",omitempty"`
	Conditions *Conditions `json:"conditions,omitempty"`
	Sequence   []*NodeLite `json:"sequence,omitempty"`
	State      *PipeState  `json:"state,omitempty"`
}

type Conditions struct {
	AlwaysTriggered bool   `json:"always_triggered,omitempty"`
	Host            string `json:"host,omitempty"`
	URLPattern      string `json:"url_pattern,omitempty"`
}

type PipesAPI struct {
	nodes            *NodesAPI
	storage          StoreInterface[*PipeRich]
	hostTriggerIndex IndexInterface
}

func (a *PipesAPI) Init(nodes *NodesAPI, storeType StorageType) error {
	*a = PipesAPI{
		nodes: nodes,
	}
	if storeType != StorageInMemory {
		return fmt.Errorf("persistent pipes not supported: %w", status.ErrUnimplemented)
	}
	a.storage = newStorageInMemory[*PipeRich]()
	a.hostTriggerIndex = newIndexInMemory()

	// Add builtin pipe.
	_, err := a.CreatePipe(&CreatePipeRequest{
		Pipe: &PipeBaseRich{
			Sequence: []*NodeBase{{
				Action: "document",
			}},
		},
	})
	return err
}

type DeletePipeRequest struct {
	ID        string `json:"id,omitempty"`
	KeepNodes bool   `json:"keep_nodes,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if !req.KeepNodes {
		pipe, err := a.storage.Load(req.ID)
		if err != nil {
			if errors.Is(err, status.ErrNotFound) {
				return &DeletePipeResponse{}, nil
			}
			return nil, err
		}
		if conds := pipe.Conditions; conds != nil {
			if host := conds.Host; host != "" {
				a.hostTriggerIndex.DeleteKeyValue(host, pipe.ID)
			}
		}
		for _, n := range pipe.Sequence {
			if _, err := a.nodes.DeleteNode(&DeleteNodeRequest{ID: n.UUID()}); err != nil && !errors.Is(err, status.ErrNotFound) {
				return nil, err
			}
		}
	}
	if _, err := a.storage.Delete(req.ID); err != nil {
		if !errors.Is(err, status.ErrNotFound) {
			return nil, err
		}
	}
	return &DeletePipeResponse{}, nil
}

type DeletePipeRequestBatch struct {
	Names []string
}

type DeletePipeResponseBatch struct{}

func (r *PipesAPI) DeleteBatch(*DeletePipeRequestBatch) (*DeletePipeResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type CreatePipeRequest struct {
	Pipe *PipeBaseRich `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	StoreOp *StorageOpLite `json:"store_op,omitempty"`
	Pipe    *PipeLite      `json:"pipe,omitempty"`
	Nodes   []*NodeLite    `json:"nodes,omitempty"`
}

func (a *PipesAPI) CreatePipe(req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) bool { _, err := a.storage.Load(id); return errors.Is(err, status.ErrNotFound) })
	if err != nil {
		return nil, err
	}
	pl := newPipeLite(id)
	seq := make([]*NodeLite, 0, len(req.Pipe.Sequence))
	for _, n := range req.Pipe.Sequence {
		id, err := newUUID(func(id string) bool { _, err := a.nodes.loadNode(id); return errors.Is(err, status.ErrNotFound) })
		if err != nil {
			return nil, err
		}
		nl := newNodeLite(id)
		node := &NodeRich{
			Node: &Node{
				NodeLite: nl,
				NodeBase: n,
			},
		}
		a.nodes.createNode(node) // createNode always returns true.
		seq = append(seq, nl)
	}
	storeOp, err := a.storage.Store(&PipeRich{
		PipeLite:   pl,
		Conditions: req.Pipe.Conditions,
		Sequence:   seq,
	})
	if err != nil {
		return nil, err
	}
	return &CreatePipeResponse{
		StoreOp: storeOp,
		Pipe:    pl,
		Nodes:   seq,
	}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []*PipeLite `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(*ListPipesRequest) (*ListPipesResponse, error) {
	n, _ := a.storage.Len()
	res := make([]*PipeLite, 0, n)
	err := a.storage.Scan(func(p *PipeRich, err error) error {
		if err != nil {
			return err
		}
		res = append(res, p.PipeLite)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ListPipesResponse{Pipes: res}, nil
}

type GetPipeRequest struct {
	Pipe string `json:"pipe,omitempty"`
}

type GetPipeResponse struct {
	Pipe *PipeRich `json:"pipe,omitempty"`
}

func (a *PipesAPI) GetPipe(req *GetPipeRequest) (*GetPipeResponse, error) {
	pipe, err := a.storage.Load(req.Pipe)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}
