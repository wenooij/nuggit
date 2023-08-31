// Package jsonglom implements merging JSON objects using the Glom semantics.
//
// See nuggit.Glom.
package jsonglom

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/keys"
)

type merge struct {
	data []byte
	glom nuggit.Glom
}

// Merges applies merges to Merge.
type MergeFunc func(srcField string) (*merge, error)

// Value applies the field of value v to the merge.
func From(v any, field string, glom nuggit.Glom) MergeFunc {
	return func(srcField string) (*merge, error) {
		v, err := Extract(field, v)
		if err != nil {
			return nil, err
		}
		data, err := render(srcField, v)
		if err != nil {
			return nil, err
		}
		return &merge{
			data: data,
			glom: glom,
		}, nil
	}
}

// Merge the field of the JSON object data with the MergeOptions.
// It returns the resulting JSON data or any merge errors.
func Merge(data json.RawMessage, field string, mergeFns ...MergeFunc) (json.RawMessage, error) {
	merges := make([]*merge, 0, len(mergeFns))
	for _, fn := range mergeFns {
		m, err := fn(field)
		if err != nil {
			return nil, err
		}
		merges = append(merges, m)
	}
	if len(data) == 0 {
		// Special case when no merges supplied.
		data = []byte("null")
	}
	for _, m := range merges {
		var err error
		data, err = mergeJSON(data, m.data, m.glom)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Extract the JSON path from the JSON marshalled any value.
func Extract(path string, v any) (any, error) {
	if path == "" {
		return v, nil
	}
	var m map[string]any
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for head, tail := keys.Cut(path); ; head, tail = keys.Cut(tail) {
		v, ok := m[head]
		if !ok {
			return nil, fmt.Errorf("key error: %q", head)
		}
		if tail == "" {
			return v, nil
		}
		switch v := v.(type) {
		case map[string]any:
			m = v
		default:
			return nil, fmt.Errorf("not a JSON object: %q", head)
		}
	}
}

func render(path string, v any) ([]byte, error) {
	if path == "" {
		return json.Marshal(v)
	}
	var buf bytes.Buffer
	for i := 0; ; i++ {
		head, tail := keys.Cut(path)
		if head == "" {
			if err := json.NewEncoder(&buf).Encode(v); err != nil {
				return nil, err
			}
			for ; i > 0; i-- {
				buf.WriteByte('}')
			}
			break
		}
		fmt.Fprintf(&buf, "{%q:", head)
		path = tail
	}
	return buf.Bytes(), nil
}

func mergeJSON(data, src []byte, glom nuggit.Glom) ([]byte, error) {
	if len(data) == 0 {
		// Special case.
		data = []byte("null")
	}
	var vd map[string]any
	if err := json.Unmarshal(data, &vd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dst: %v", err)
	}
	if vd == nil {
		// Special case to handle "null".
		vd = make(map[string]any)
	}
	var vs map[string]any
	if err := json.Unmarshal(src, &vs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal src: %v", err)
	}
	if err := mergeObjects(vd, vs, glom); err != nil {
		return nil, err
	}
	return json.Marshal(vd)
}

func mergeObjects(dst, src map[string]any, glom nuggit.Glom) error {
	for k, v := range src {
		if err := doGlom(dst, k, v, glom); err != nil {
			return err
		}
	}
	return nil
}

func doGlom(dst map[string]any, k string, v any, glom nuggit.Glom) error {
	dv, ok := dst[k]
	if !ok {
		dst[k] = v
		return nil
	}
	switch dv := dv.(type) {
	case map[string]any:
		switch v := v.(type) {
		case map[string]any:
			return mergeObjects(dv, v, glom)
		default:
			return fmt.Errorf("merge error: merging object with non-object: %T", v)
		}
	case string:
		switch glom {
		case nuggit.GlomAppend, nuggit.GlomExtend:
			dst[k] = fmt.Sprint(dv, v)
		case nuggit.GlomUndefined, nuggit.GlomAssign:
			dst[k] = dv
		default:
			return fmt.Errorf("glom error: unknown glom: %v", glom)
		}
	case []any:
		switch glom {
		case nuggit.GlomAppend:
			dst[k] = append(dv, v)
		case nuggit.GlomExtend:
			switch v := v.(type) {
			case []any:
				dst[k] = append(dv, v...)
			default:
				return fmt.Errorf("glom error: extend not defined for: (%T, %T)", dv, v)
			}
		case nuggit.GlomUndefined, nuggit.GlomAssign:
			dst[k] = v
		default:
			return fmt.Errorf("glom error: unknown glom: %v", glom)
		}
	default:
		switch glom {
		case nuggit.GlomUndefined, nuggit.GlomAssign:
			dst[k] = v
		case nuggit.GlomAppend, nuggit.GlomExtend:
			return fmt.Errorf("glom not defined for lhs: %T", dv)
		}
	}
	return nil
}
