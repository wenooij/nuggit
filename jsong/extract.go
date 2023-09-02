package jsong

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wenooij/nuggit/keys"
)

// Extract the JSON path from the JSON marshalled any value.
// Path components comprise elements each seperated by a dot.
// Path elements may be field names (by the JSON marshaled map key)
// or array indices for arrays and slices (as a 0-based integers).
// The definition of the key format is in nuggit.FieldKey.
func Extract(v any, path string) (any, error) {
	return reflectExtract(reflect.ValueOf(v), path)
}

func reflectExtract(rv reflect.Value, path string) (v any, err error) {
	if path == "" {
		return reflectValueOf(rv)
	}
	head, tail, leaf := keys.Cut(path)
	if head == "" {
		return nil, fmt.Errorf("empty elem in key: %q", path)
	}
	if !leaf && tail == "" {
		return nil, fmt.Errorf("empty tail in path: %q", path)
	}
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Array:
		i, ok := keys.Index(head)
		if !ok {
			return nil, fmt.Errorf("not an index: %q", head)
		}
		if int64(rv.Len()) <= i {
			return nil, fmt.Errorf("array index %d out of bounds: %d", i, rv.Len())
		}
		return reflectExtract(rv.Index(int(i)), tail)
	case reflect.Slice:
		i, ok := keys.Index(head)
		if !ok {
			return nil, fmt.Errorf("not an index: %q", head)
		}
		if int64(rv.Len()) <= i {
			return nil, fmt.Errorf("array index %d out of bounds: %d", i, rv.Len())
		}
		return reflectExtract(rv.Index(int(i)), tail)
	case reflect.Map:
		var k any
		switch rv.Type().Key().Kind() {
		case reflect.String:
			k = head
		case reflect.Int:
			i, ok := keys.Index(head)
			if !ok {
				return nil, fmt.Errorf("not an int: %q", head)
			}
			k = i
		default:
			return nil, fmt.Errorf("unsupported key Kind for map: %v", k)
		}
		rv := rv.MapIndex(reflect.ValueOf(k))
		return reflectExtract(rv, tail)
	case reflect.Struct:
		t := rv.Type()
		f, ok := t.FieldByNameFunc(func(name string) bool {
			f, _ := t.FieldByName(name)
			if jtag, ok := f.Tag.Lookup("json"); ok {
				name = strings.TrimSuffix(jtag, ",omitempty")
			}
			if name != "-" && name == head {
				return true
			}
			return false
		})
		if !ok {
			return nil, fmt.Errorf("no key in struct: %q", head)
		}
		return reflectExtract(rv.FieldByIndex(f.Index), tail)
	default:
		return nil, fmt.Errorf("unsupported Kind: %v", rv.Kind())
	}
}

func fallbackExtract(x any, path string) (any, error) {
	v, err := ValueOf(x)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return v, nil
	}
	for head, tail, _ := keys.Cut(path); ; head, tail, _ = keys.Cut(tail) {
		var t any
		switch v := v.(type) {
		case []any:
			i, ok := keys.Index(head)
			if !ok {
				return nil, fmt.Errorf("not an array index: %q", head)
			}
			if int64(len(v)) <= i {
				return nil, fmt.Errorf("array index %d out of bounds: %d", i, len(v))
			}
			t = v[i]
		case map[string]any:
			var ok bool
			t, ok = v[head]
			if !ok {
				return nil, fmt.Errorf("key error: %q", head)
			}
		default:
			return nil, fmt.Errorf("not a JSON object or array at %q: %T", head, v)
		}
		if tail == "" {
			return t, nil
		}
		v = t
	}
}
