package api

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestResourceUnmarshalYAML(t *testing.T) {
	data := []byte(`api_version: v1
kind: pipe
metadata:
  name: foo
spec:
  actions:
    - action: selector
      spec:
        selector: div.foo`)
	r := new(Resource)
	if err := yaml.Unmarshal(data, r); err != nil {
		t.Fatal(err)
	}
	data, _ = json.MarshalIndent(r, "", "  ")
	t.Log(string(data))
}
