package nuggit

type View struct {
	Alias      string       `json:"alias,omitempty"`
	Columns    []ViewColumn `json:"columns,omitempty"`
	AggColumns []AggColumn  `json:"agg_columns,omitempty"`
}

func (v *View) GetSpec() any { return v }

func (v *View) GetAlias() string {
	if v == nil {
		return ""
	}
	return v.Alias
}

func (v *View) GetColumns() []ViewColumn {
	if v == nil {
		return nil
	}
	return v.Columns
}

type ViewColumn struct {
	Alias string `json:"alias,omitempty"`
	Pipe  string `json:"pipe,omitempty"`
	Point Point  `json:"point,omitempty"`
}

type AggColumn struct {
	ViewColumn `json:",omitempty"`
	Op         AggOp  `json:"op,omitempty"`
	Arg        string `json:"arg,omitempty"`
	Distinct   bool   `json:"distinct,omitempty"`
	Expr       string `json:"expr,omitempty"`
	Filter     string `json:"filter,omitempty"`
	OrderBy    string `json:"order_by,omitempty"`
}

type AggOp = string

const (
	AggOpCount     = "count"
	AggOpSum       = "sum"
	AggOpMin       = "min"
	AggOpMax       = "max"
	AggOpAvg       = "avg"
	AggOpStringAgg = "string_agg"
)
