package v1alpha

import (
	"encoding/json"
	"fmt"
)

type StringData struct {
	String string   `json:"string,omitempty"`
	Op     StringOp `json:"op,omitempty"`
	Sep    string   `json:"sep,omitempty"`
	Begin  int      `json:"begin,omitempty"`
	End    int      `json:"end,omitempty"`
}

type String struct {
	StringData `json:",omitempty"`
}

func Bind([]Edge) error {
	return nil
}

func (s *String) UnmarshalJSON(data []byte) error {
	var sd StringData
	if err := json.Unmarshal(data, &sd); err == nil {
		s.StringData = sd
		return nil
	}
	return fmt.Errorf("failed to unmarshal String")
}
