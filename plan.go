package nuggit

type Plan struct {
	Stages    []Graph    `json:"stages,omitempty"`
	Exchanges []Exchange `json:"exchanges,omitempty"`
}

// Exchange is an Edge that crosses Stage boundaries.
// It adds the SrcGraph
type Exchange struct {
	SrcStage StageKey `json:"src_stage,omitempty"`
	// DstGraph
	DstStage StageKey `json:"dst_stage,omitempty"`
	// Edge represents the logical connection where Src is a Node in SrcGraph
	// and Dst is a Node in DstGraph.
	Edge `json:",omitempty"`
}
