package v1alpha

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

type Find struct {
	Literal string `json:"literal,omitempty"`
	Byte    *byte  `json:"byte,omitempty"`
	AnyByte []byte `json:"any_byte,omitempty"`
	AnyRune string `json:"any_rune,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	// Regex is a regular expression pattern.
	Regex *Regex `json:"regex,omitempty"`
	Sink  *Sink  `json:"sink,omitempty"`
	// All marks whether to find
	All bool `json:"all,omitempty"`
	// Submatch marks whether to include matching groups in the results.
	Submatch      bool   `json:"submatch,omitempty"`
	Index         bool   `json:"index,omitempty"`
	SubmatchIndex int    `json:"submatch_index,omitempty"`
	GroupName     string `json:"group_name,omitempty"`
	// Reverse marks that the search be conducted in reverse.
	Reverse bool `json:"reverse,omitempty"`
}

func (x *Find) Run(ctx context.Context) (any, error) {
	if x.Sink == nil {
		return nil, fmt.Errorf("Find must have a Sink")
	}
	if x.Offset < 0 {
		return nil, fmt.Errorf("invalid Offset < 0")
	}
	var data []byte
	{
		result, err := x.Sink.Run(ctx)
		if err != nil {
			return nil, err
		}
		data = result.([]byte)
	}
	if len(data) < x.Offset {
		return nil, fmt.Errorf("offset out of bounds: len = %v, offset = %v", len(data), x.Offset)
	}
	if x.Offset > 0 {
		data = data[x.Offset:]
	}
	if x.Regex == nil {
		return x.runNoRegex(data)
	}
	if x.Reverse {
		// TODO(wes): Support reverse.
		return nil, fmt.Errorf("setting Reverse is not supported for Regex")
	}
	var re *regexp.Regexp
	{
		result, err := x.Regex.Run(ctx)
		if err != nil {
			return nil, err
		}
		re = result.(*regexp.Regexp)
	}
	if x.Reverse {
		slices.Reverse(data)
	}
	switch {
	case x.All && x.Submatch:
		return re.FindAllSubmatchIndex(data, -1), nil
	case x.All:
		return re.FindAllIndex(data, -1), nil
	case x.Submatch:
		return re.FindSubmatchIndex(data), nil
	default:
		return re.FindIndex(data), nil
	}
}

func (x *Find) runNoRegex(data []byte) (any, error) {
	switch {
	case x.Literal != "":
		if x.Reverse {
			return bytes.LastIndex(data, []byte(x.Literal)), nil
		}
		return bytes.Index(data, []byte(x.Literal)), nil
	case x.Byte != nil:
		if x.Reverse {
			return bytes.LastIndexByte(data, *x.Byte), nil
		}
		return bytes.IndexByte(data, *x.Byte), nil
	case len(x.AnyByte) > 0:
		if x.Reverse {
			return bytes.LastIndexAny(data, string(x.AnyByte)), nil
		}
		return bytes.IndexAny(data, string(x.AnyByte)), nil
	case x.AnyRune != "":
		if x.Reverse {
			return strings.LastIndexAny(string(data), x.AnyRune), nil
		}
		return strings.IndexAny(string(data), x.AnyRune), nil
	default:
		return nil, fmt.Errorf("no conditions set")
	}
}
