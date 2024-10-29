package nuggit

type View struct {
	Alias   string       `json:"alias,omitempty"`
	Columns []ViewColumn `json:"columns,omitempty"`
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
