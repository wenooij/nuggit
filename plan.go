package nuggit

// Plan represents a multistage execution between a number of graphs.
// Graphs are executed in concurrent stages with data exchanged
// in accordance with the Exchange edges.
type Plan struct {
	Graphs    []Graph    `json:"graphs,omitempty"`
	Exchanges []Exchange `json:"exchanges,omitempty"`
}

// Exchange is an Edge that crosses Graph boundaries.
// It adds the SrcStage and DstStage fields.
type Exchange struct {
	SrcStage StageKey `json:"src_stage,omitempty"`
	// DstGraph
	DstStage StageKey `json:"dst_stage,omitempty"`
	// Edge represents the logical connection where Src is a Node in SrcGraph
	// and Dst is a Node in DstGraph.
	Edge `json:",omitempty"`
}
