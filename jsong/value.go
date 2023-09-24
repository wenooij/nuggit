package jsong

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ValueOf creates the jsong value of the input v.
// ValueOf performs an operation equivalent to but more efficient than
// a MarshalJSON followed by an Unmarshal to any.
func ValueOf(v any) (any, error) {
	if v, ok := fastValueOf(v); ok {
		return v, nil
	}
	return reflectValueOf(reflect.ValueOf(v))
}

func fastValueOf(v any) (any, bool) {
	if v == any(nil) {
		return nil, true
	}
	switch v := v.(type) {
	case bool, float64, string:
		return v, true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint:
		return float64(v), true
	case float32:
		return float64(v), true
	case []byte:
		return string(v), true
	default:
		return nil, false
	}
}

func reflectValueOf(rv reflect.Value) (any, error) {
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if v, ok := fastValueOf(rv.Interface()); ok {
		return v, nil
	}
	switch rv.Kind() {
	case reflect.Bool:
		var out bool
		reflect.ValueOf(&out).Elem().Set(rv.Convert(boolType))
		return out, nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Float32, reflect.Float64:
		var out float64
		reflect.ValueOf(&out).Elem().Set(rv.Convert(float64Type))
		return out, nil
	case reflect.String:
		var out string
		reflect.ValueOf(&out).Elem().Set(rv.Convert(stringType))
		return out, nil
	case reflect.Array:
		vCopy := make([]any, rv.Len(), rv.Cap())
		for i := 0; i < rv.Len(); i++ {
			e, err := ValueOf(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			vCopy[i] = e
		}
		return vCopy, nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// Special case for []byte.
			var out string
			reflect.ValueOf(&out).Elem().Set(rv.Convert(stringType))
			return out, nil
		}
		vCopy := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			e, err := ValueOf(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			vCopy[i] = e
		}
		return vCopy, nil
	case reflect.Map:
		if k := rv.Type().Key().Kind(); k != reflect.String {
			return nil, fmt.Errorf("unsupported key Kind for map: %v", k)
		}
		st := reflect.TypeOf("")
		vCopy := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key().Convert(st)
			v, err := ValueOf(iter.Value().Interface())
			if err != nil {
				return nil, fmt.Errorf("in map key %v: %v", k, err)
			}
			vCopy[k.Interface().(string)] = v
		}
		return vCopy, nil
	case reflect.Struct:
		t := rv.Type()
		vCopy := make(map[string]any, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := f.Name
			var omitEmpty bool
			if jtag, ok := f.Tag.Lookup("json"); ok {
				omitEmpty = strings.HasSuffix(jtag, ",omitempty")
				name = strings.TrimSuffix(jtag, ",omitempty")
			}
			if name == "-" {
				continue // Skip no JSON.
			}
			v, err := ValueOf(rv.Field(i).Interface())
			if err != nil {
				return nil, err
			}
			if name == "" && f.Anonymous {
				vm := v.(map[string]any)
				// Embed resulting values into the map directly
				// While avoiding collisions.
				var toDeleteKeys []string
				for k, v := range vm {
					if _, ok := vCopy[k]; !ok {
						vCopy[k] = v
						toDeleteKeys = append(toDeleteKeys, k)
					}
				}
				for _, k := range toDeleteKeys {
					delete(vm, k)
				}
				if len(vm) > 0 {
					// Any collision keys are stored in the original place.
					vCopy[f.Name] = vm
				}
				continue
			}
			if !omitEmpty || !reflect.ValueOf(v).IsZero() {
				vCopy[name] = v
			}
		}
		return vCopy, nil
	default:
		return nil, fmt.Errorf("unsupported Kind: %v", rv.Kind())
	}
}

func valueOfSlice[T ~[]E, E any](slice T) (any, error) {
	var e *E
	if _, ok := any(e).(*uint8); ok {
		// Special case for []byte.
		return string(any(slice).([]byte)), nil
	}
	out := make([]any, len(slice))
	for i, v := range slice {
		e, err := ValueOf(v)
		if err != nil {
			return nil, err
		}
		out[i] = e
	}
	return out, nil
}

func fallbackValueOf(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var res any
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}
