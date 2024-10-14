package api

type CollectionLite struct {
	*Ref `json:",omitempty"`
}

type Collection struct {
	*Ref `json:",omitempty"`
	Name string `json:"name,omitempty"`
}

func (c *Collection) Lite() *CollectionLite {
	if c == nil {
		return nil
	}
	return &CollectionLite{c.Ref}
}

type PointLite struct {
	*Ref `json:",omitempty"`
}

type PointBase struct {
	Type       Type `json:"type,omitempty"`
	Collection *Ref `json:"collection,omitempty"`
}

type Point struct {
	*PointLite `json:",omitempty"`
	*PointBase `json:",omitempty"`
}
