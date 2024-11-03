//go:build wasm

package main

import (
	"fmt"
	"regexp"
	"strings"
	"syscall/js"
)

type Action interface {
	Execute(input js.Value) js.Value
}

type MapAction struct {
	action string
	mapper func(js.Value) js.Value
}

func (a MapAction) Execute(input js.Value) js.Value {
	return value_asArray(input).Call("map", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
		return a.mapper(args[0])
	})))
}

type FilterAction struct {
	MapAction
	filter func(js.Value) bool
}

func (a FilterAction) Execute(input js.Value) js.Value {
	return value_asArray(input).Call("flatMap", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
		e := args[0]
		if a.filter(e) {
			return []any{a.mapper(e)}
		}
		return []any{}
	})))
}

// TODO: Create a version of this that doesn't flattens the value batch.
func PropAction(prop string) MapAction {
	return MapAction{
		action: "prop",
		mapper: func(e js.Value) js.Value {
			// null
			// undefined
			if e.IsNull() || e.IsUndefined() {
				return e
			}
			// [object]
			// string
			switch e.Type() {
			case js.TypeObject:
				return e.Get(prop)
			default:
				return js.Null()
			}
		},
	}
}

type DocumentElementAction struct{}

func (a DocumentElementAction) Execute(_ js.Value) js.Value {
	return js.Global().Get("document").Get("documentElement")
}

func FilterSelectorAction(selector string) FilterAction {
	// https://developer.mozilla.org/docs/Web/API/Element
	element := js.Global().Get("Element")
	sel := js.ValueOf(selector)
	return FilterAction{
		MapAction: MapAction{
			action: "filterSelector",
		},
		filter: func(e js.Value) bool {
			return e.InstanceOf(element) && e.Call("matches", sel).Bool()
		},
	}
}

func QuerySelectorAction(selector string, all, self bool) MapAction {
	matchSelf := FilterSelectorAction(selector)
	// https://developer.mozilla.org/docs/Web/API/Element
	element := js.Global().Get("Element")
	return MapAction{
		action: "querySelector",
		mapper: func(e js.Value) js.Value {
			if !e.InstanceOf(element) {
				return js.Null()
			}
			matches := js.ValueOf([]any{})
			if self && matchSelf.filter(e) {
				matches.Call("push", e)
			}
			sel := js.ValueOf(selector)
			var method string
			if all {
				method = "querySelectorAll"
			} else {
				method = "querySelector"
			}
			return matches.Call("concat", e.Call(method, sel))
		},
	}
}

func RegexpAction(pattern string) Action {
	re := regexp.MustCompile(pattern)
	return MapAction{
		action: "regexp",
		mapper: func(e js.Value) js.Value {
			var matches []any
			s := e.String()
			for _, m := range re.FindAllStringSubmatch(s, -1) {
				if len(m) > 1 { // Use first group if available.
					matches = append(matches, m[1])
				} else { // Fall back to full match.
					matches = append(matches, m[0])
				}
			}
			return js.ValueOf(matches)
		},
	}
}

func SplitAction(separator string) Action {
	return MapAction{
		action: "split",
		mapper: func(e js.Value) js.Value {
			ss := strings.Split(e.String(), separator)
			res := make([]any, len(ss))
			for i, v := range ss {
				res[i] = v
			}
			return js.ValueOf(res)
		},
	}
}

type Chain []Action

func (c Chain) Execute(input js.Value) js.Value {
	res := input
	for _, a := range c {
		res = a.Execute(res)
	}
	return res
}

func CreateAction(config js.Value) (Action, error) {
	action := config.Get("action").String()
	switch action {
	case "documentElement": // https://developer.mozilla.org/en-US/docs/Web/API/Document/documentElement
		return DocumentElementAction{}, nil
	case "filterSelector": // https://developer.mozilla.org/en-US/docs/Web/API/Element/matches
		return FilterSelectorAction(config.Get("selector").String()), nil
	case "querySelector": // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelector
		return QuerySelectorAction(
			config.Get("selector").String(),
			config.Get("self").Bool(),
			config.Get("all").Bool(),
		), nil
	case "innerHTML": // https://developer.mozilla.org/en-US/docs/Web/API/Element/innerHTML
		return PropAction("innerHTML"), nil
	case "outerHTML": // https://developer.mozilla.org/en-US/docs/Web/API/Element/outerHTML
		return PropAction("outerHTML"), nil
	case "innerText": // https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/innerText
		return PropAction("innerText"), nil
	case "regexp": // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp
		return RegexpAction(config.Get("pattern").String()), nil
	case "attributes": // https://developer.mozilla.org/en-US/docs/Web/API/Element/attributes
		return Chain{PropAction(config.Get("attributes").String()), PropAction(config.Get("name").String())}, nil
	case "split": // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/split
		return SplitAction(config.Get("separator").String()), nil
	case "get": // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/get#prop
		return PropAction(config.Get("prop").String()), nil
	default:
		return nil, fmt.Errorf("unsupported action (%q)", action)
	}
}
