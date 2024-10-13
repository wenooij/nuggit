package runtime

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/client"
	"github.com/wenooij/nuggit/status"
)

type RunRequest struct {
	Pipeline string      `json:"pipeline,omitempty"`
	Args     client.Args `json:"args,omitempty"`
	Data     []byte      `json:"data,omitempty"`
}

type RunResponse struct{}

func (r *Runtime) Run(*RunRequest) (*RunResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type RunRequestBatch struct {
	Args []client.Args
}

type RunResponseBatch struct{}

func (r *Runtime) RunBatch(*RunRequestBatch) (*RunResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type EnableRequest struct {
	Name    string
	Enabled bool
}

type EnableResponse struct{}

func (r *Runtime) Enable(*EnableRequest) (*EnableResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type DeletePipelineRequest struct {
	ID        string `json:"id,omitempty"`
	KeepNodes bool   `json:"keep_nodes,omitempty"`
}

type DeletePipelineResponse struct{}

func (r *Runtime) DeletePipeline(req *DeletePipelineRequest) (*DeletePipelineResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deletePipeline(req.ID, req.KeepNodes)
	return &DeletePipelineResponse{}, nil
}

type DeletePipelineRequestBatch struct {
	Names []string
}

type DeletePipelineResponseBatch struct{}

func (r *Runtime) DeleteBatch(*DeletePipelineRequestBatch) (*DeletePipelineResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type PutPipelineRequest struct {
	Pipeline *PipelineWithNodes `json:"pipeline,omitempty"`
}

type PutPipelineResponse struct {
	Pipeline ListPipelinesPipeline `json:"pipeline,omitempty"`
}

type PipelineWithNodes struct {
	RunCondition nuggit.RunCondition `json:"cond,omitempty"`
	Sequence     []ListNode          `json:"sequence,omitempty"`
}

func Provided[T comparable](arg string, t T) error {
	var zero T
	if t == zero {
		return fmt.Errorf("%s is required: %w", arg, status.ErrInvalidArgument)
	}
	return nil
}

func (r *Runtime) PutPipeline(req *PutPipelineRequest) (*PutPipelineResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, err := NewUUID(func(id string) bool { return r.pipelines[id] == nil })
	if err != nil {
		return nil, err
	}
	seq := make([]string, 0, len(req.Pipeline.Sequence))
	for i, n := range req.Pipeline.Sequence {
		if n.ID != "" && n.Node != nil {
			return nil, fmt.Errorf("provide either id or node but not both (%v): %w", n.ID, status.ErrInvalidArgument)
		} else if n.ID != "" {
			node := r.nodes[n.ID]
			if node == nil {
				return nil, fmt.Errorf("node not found (%v): %w", n.ID, status.ErrNotFound)
			}
			seq = append(seq, n.ID)
		} else if n.Node != nil {
			// Create this node.
			id, err := NewUUID(func(id string) bool { return r.nodes[id] == nil })
			if err != nil {
				return nil, err
			}
			r.putNode(id, n.Node)
		} else {
			return nil, fmt.Errorf("provide either id or node (#%d): %w", i, status.ErrInvalidArgument)
		}
	}
	pipe := &nuggit.Pipeline{
		RunCondition: &req.Pipeline.RunCondition,
		Sequence:     seq,
	}
	r.pipelines[id] = pipe
	return &PutPipelineResponse{Pipeline: MakeListPipelinesPipeline(id, pipe)}, nil
}

type ListPipelinesRequest struct{}

type ListPipelinesPipeline struct {
	ID       string              `json:"id,omitempty"`
	Self     string              `json:"self,omitempty"`
	Sequence []ListPipelinesNode `json:"sequence,omitempty"`
}

type ListPipelinesNode struct {
	ID  string `json:"id,omitempty"`
	URI string `json:"uri,omitempty"`
}

type ListPipelinesResponse struct {
	Pipelines map[string]ListPipelinesPipeline `json:"pipelines,omitempty"`
}

func (r *Runtime) ListPipelines(*ListPipelinesRequest) (*ListPipelinesResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make(map[string]ListPipelinesPipeline, len(r.pipelines))
	for id, p := range r.pipelines {
		res[id] = MakeListPipelinesPipeline(id, p)
	}
	return &ListPipelinesResponse{Pipelines: res}, nil
}

func MakeListPipelinesPipeline(id string, p *nuggit.Pipeline) ListPipelinesPipeline {
	res := ListPipelinesPipeline{
		ID:   id,
		Self: fmt.Sprintf("/api/pipelines/%s", id),
	}
	for _, id := range p.Sequence {
		res.Sequence = append(res.Sequence, ListPipelinesNode{
			ID:  id,
			URI: fmt.Sprintf("/api/nodes/%s", id),
		})
	}
	return res
}

type ListPipelineRequest struct {
	Pipeline string `json:"pipeline,omitempty"`
}

type ListPipelineResponse struct {
	Pipeline ListPipelinesPipeline `json:"pipeline,omitempty"`
}

func (r *Runtime) ListPipeline(req *ListPipelineRequest) (*ListPipelineResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pipe, ok := r.pipelines[req.Pipeline]
	if !ok {
		return nil, fmt.Errorf("failed to load pipeline: %w", status.ErrNotFound)
	}
	return &ListPipelineResponse{Pipeline: MakeListPipelinesPipeline(req.Pipeline, pipe)}, nil
}

type ListNodesRequest struct{}

type ListNodesResponse struct {
	Nodes map[string]nuggit.RawNode `json:"nodes,omitempty"`
}

func (r *Runtime) ListNodes(*ListNodesRequest) (*ListNodesResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make(map[string]nuggit.RawNode, len(r.nodes))
	for id, node := range r.nodes {
		res[id] = *node
	}
	return &ListNodesResponse{Nodes: res}, nil
}

type ListNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type ListNode struct {
	ID   string          `json:"id,omitempty"`
	Self string          `json:"self,omitempty"`
	Node *nuggit.RawNode `json:"node,omitempty"`
}

type ListNodeResponse struct {
	Node ListNode `json:"node,omitempty"`
}

func sameNodes(a, b *nuggit.RawNode) bool {
	return a == nil && b == nil ||
		a != nil && b != nil && a.Action == b.Action && bytes.Equal(a.Spec, b.Spec)
}

func (r *Runtime) ListNode(req *ListNodeRequest) (*ListNodeResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	node, ok := r.nodes[req.ID]
	if !ok {
		return nil, fmt.Errorf("failed to load node: %w", status.ErrNotFound)
	}
	return &ListNodeResponse{Node: ListNode{
		ID:   req.ID,
		Self: fmt.Sprintf("/api/nodes/%s", req.ID),
		Node: node,
	}}, nil
}

type ListNodeUsesRequest struct {
	ID string `json:"id,omitempty"`
}

type ListNodeUsesResponse struct {
	Pipelines []ListPipelinesPipeline `json:"pipelines,omitempty"`
}

func (r *Runtime) ListNodeUses(req *ListNodeUsesRequest) (*ListNodeUsesResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pipelines, ok := r.nodeUses[req.ID]
	if !ok {
		return nil, fmt.Errorf("failed to load node: %w", status.ErrNotFound)
	}
	resp := make([]ListPipelinesPipeline, 0, len(pipelines))
	for id := range pipelines {
		resp = append(resp, ListPipelinesPipeline{
			ID:   id,
			Self: fmt.Sprintf("/api/pipelines/%s", id),
		})
	}
	return &ListNodeUsesResponse{Pipelines: resp}, nil
}

type DeleteNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteNodeResponse struct{}

func (r *Runtime) DeleteNode(req *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := Provided("id", req.ID); err != nil {
		return nil, err
	}
	if err := ValidateUUID(req.ID); err != nil {
		return nil, err
	}
	if r.nodes[req.ID] != nil && !r.deleteNode(req.ID) {
		return nil, fmt.Errorf("node is in use by at least one pipeline: %w", status.ErrFailedPrecondition)
	}
	return &DeleteNodeResponse{}, nil
}

type TriggerRequest struct {
	URL string `json:"url,omitempty"`
}

type TriggerResponse struct {
	Roots []string                  `json:"roots,omitempty"`
	Nodes map[string]nuggit.RawNode `json:"nodes,omitempty"`
}

func (r *Runtime) Trigger(req *TriggerRequest) (*TriggerResponse, error) {
	roots := make([]string, 0)
	nodes := make(map[string]nuggit.RawNode)
	for _, p := range r.alwaysOnPipelines {
		if root, ok := p.Root(); ok {
			roots = append(roots, root)
		}
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	for _, p := range r.pipelinesByHost[u.Hostname()] {
		if root, ok := p.Root(); ok {
			roots = append(roots, root)
		}
	}
	return &TriggerResponse{Roots: roots, Nodes: nodes}, nil
}

type StatusRequest struct{}

type StatusResponse struct{}

func (r *Runtime) Status(*StatusRequest) (*StatusResponse, error) { return &StatusResponse{}, nil }

type PutNodeRequest struct {
	Node *nuggit.RawNode `json:"node,omitempty"`
}

type PutNodeResponse struct {
	Node ListNode `json:"node,omitempty"`
}

func (r *Runtime) PutNode(req *PutNodeRequest) (*PutNodeResponse, error) {
	if err := Provided("node", req.Node); err != nil {
		return nil, err
	}
	// Generate an ID while avoiding extremely rare conflicts.
	var id string
	for {
		u, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		id = u.String()
		if r.nodes[id] == nil {
			break
		}
	}
	r.putNode(id, req.Node)
	return &PutNodeResponse{
		Node: ListNode{
			ID:   id,
			Self: fmt.Sprintf("/api/nodes/%s", id),
		},
	}, nil
}

type ListOrphansRequest struct{}

type ListOrphansResponse struct {
	Nodes []ListNode `json:"nodes,omitempty"`
}

func (r *Runtime) ListOrphans(*ListOrphansRequest) (*ListOrphansResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	orphans := make([]ListNode, 0, len(r.nodeUses))
	for id, uses := range r.nodeUses {
		if len(uses) == 0 {
			orphans = append(orphans, ListNode{
				ID:   id,
				Self: fmt.Sprintf("/api/nodes/%s", id),
			})
		}
	}
	return &ListOrphansResponse{Nodes: orphans}, nil
}

type DeleteOrphansRequest struct{}

type DeleteOrphansResponse struct{}

func (r *Runtime) DeleteOrphans(*DeleteOrphansRequest) (*DeleteOrphansResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	orphans := make([]string, 0, len(r.nodeUses))
	for id, uses := range r.nodeUses {
		if len(uses) == 0 {
			orphans = append(orphans, id)
		}
	}
	for _, id := range orphans {
		r.deleteNode(id)
	}
	return &DeleteOrphansResponse{}, nil
}

func NewUUID(uniqueCheck func(id string) bool) (string, error) {
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

func ValidateUUID(s string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	return nil
}
