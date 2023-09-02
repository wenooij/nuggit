package jsong

import (
	"fmt"

	"github.com/wenooij/nuggit/keys"
)

// Merge the field of the JSON object data with the MergeOptions.
// It returns the resulting JSON data or any merge errors.
func Merge(dst, src any, dstField, srcField string) (any, error) {
	src, err := Extract(src, srcField)
	if err != nil {
		return nil, fmt.Errorf("failed to extract src field: %w", err)
	}
	dstVal, err := ValueOf(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to convert dst to value: %w", err)
	}
	if err := recMerge(dstVal, src, dstField); err != nil {
		return nil, fmt.Errorf("failed to merge: %w", err)
	}
	return dstVal, nil
}

func recMerge(dst, src any, dstField string) error {
	if dstField == "" {
		return nil
	}
	head, tail, leaf := keys.Cut(dstField)
	if head == "" {
		return fmt.Errorf("empty elem in key")
	}
	if !leaf && tail == "" {
		return fmt.Errorf("empty tail in key")
	}
	switch dst := dst.(type) {
	case []any:
		i, ok := keys.Index(head)
		if !ok {
			return fmt.Errorf("not an index: %q", head)
		}
		if int64(len(dst)) <= i {
			return fmt.Errorf("array index %d out of bounds: %d", i, len(dst))
		}
		if tail == "" {
			dst[i] = src
			return nil
		}
		return recMerge(dst[i], src, tail)
	case map[string]any:
		if tail == "" {
			dst[head] = src
			return nil
		}
		return recMerge(dst[head], src, tail)
	default:
		return fmt.Errorf("unsupported type: %T", dst)
	}
}
