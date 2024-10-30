package nuggit

type Package struct {
	Rules            map[string]Rule     `json:"rules,omitempty"`
	Views            map[string]View     `json:"views,omitempty"`
	Pipes            map[string]Pipe     `json:"pipes,omtiempty"`
	AdditionalLabels map[string][]string `json:"additional_labels,omitempty"`
}

func (p *Package) GetRules() map[string]Rule {
	if p == nil {
		return nil
	}
	return p.Rules
}

func (p *Package) GetViews() map[string]View {
	if p == nil {
		return nil
	}
	return p.Views
}

func (p *Package) GetPipes() map[string]Pipe {
	if p == nil {
		return nil
	}
	return p.Pipes
}

func (p *Package) GetAdditionalLabels() map[string][]string {
	if p == nil {
		return nil
	}
	return p.AdditionalLabels
}
