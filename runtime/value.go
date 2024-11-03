//go:build wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/wenooij/nuggit"
)

func value_isString(v js.Value) bool {
	return v.Type() == js.TypeString
}

func value_isNumber(v js.Value) bool {
	return v.Type() == js.TypeNumber
}

func value_isArray(v js.Value) bool {
	return js.Global().Get("Array").Call("isArray", v).Bool()
}

var nodeList = js.Global().Get("NodeList")

func value_isNodeList(v js.Value) bool {
	return v.InstanceOf(nodeList)
}

var htmlElement = js.Global().Get("HTMLElement")

// https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement
func value_isHTMLElement(v js.Value) bool {
	return v.InstanceOf(htmlElement)
}

var element = js.Global().Get("Element")

// https://developer.mozilla.org/en-US/docs/Web/API/Element
func value_isElement(v js.Value) bool {
	return v.InstanceOf(element)
}

var node = js.Global().Get("Node")

// https://developer.mozilla.org/en-US/docs/Web/API/Node
func value_isNode(v js.Value) bool {
	return v.InstanceOf(node)
}

var attr = js.Global().Get("Attr")

// https://developer.mozilla.org/en-US/docs/Web/API/Attr
func value_isAttr(v js.Value) bool {
	return v.InstanceOf(attr)
}

var namedNodeMap = js.Global().Get("NamedNodeMap")

// https://developer.mozilla.org/en-US/docs/Web/API/NamedNodeMap
func value_isNamedNodeMap(v js.Value) bool {
	return v.InstanceOf(namedNodeMap)
}

func value_asArray(v js.Value) js.Value {
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array
	if value_isArray(v) {
		return v
	}
	// https://developer.mozilla.org/en-US/docs/Web/API/NodeList
	if value_isNodeList(v) {
		return js.Global().Get("Array").Call("from", v)
	}
	return js.Global().Get("Array").Call("of", v)
}

// normalize the value by converting applicable parts to arrays.
func value_normalize(v js.Value) js.Value {
	if v.IsNull() || v.IsUndefined() {
		return js.Null()
	}
	if value_isString(v) || value_isNumber(v) {
		return v
	}
	if value_isArray(v) {
		v.Call("map", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
			return value_normalize(args[0])
		})))
	}
	if value_isElement(v) {
		return v.Get("outerHTML")
	}
	if value_isNode(v) {
		// TODO: Do we have something more appropriate to return?
		// See https://developer.mozilla.org/en-US/docs/Web/API/Node/textContent#differences_from_innertext
		return v.Get("textContent")
	}
	if value_isAttr(v) {
		return js.ValueOf(fmt.Sprintf("%s=%q", v.Get("name").String(), v.Get("value").String()))
	}
	if value_isNamedNodeMap(v) || value_isNodeList(v) {
		return value_normalize(js.Global().Get("Array").Call("from", v))
	}
	// unexpected value in normalized will be returned as is
	return v
}

// cast casts the normalized value to the value expected by the given scalar.
//
// It returns the casted value.
func value_cast(v js.Value, scalar nuggit.Scalar) js.Value {
	// The scalar value is batched, cast it pointwise.
	if value_isArray(v) {
		return v.Call("map", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
			return value_castScalar(args[0], scalar)
		})))
	}
	return value_castScalar(v, scalar)
}

func value_castRepeated(v js.Value, scalar nuggit.Scalar) js.Value {
	// Repeated array values are cast pointwise using map.
	if value_isArray(v) {
		return v.Call("map", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
			return value_cast(args[0], scalar)
		})))
	}
	// Wrap the casted scalar result in an array.
	return js.Global().Get("Array").Call("of", value_castScalar(v, scalar))
}

func value_castScalar(v js.Value, scalar nuggit.Scalar) js.Value {
	// Regardless of the scalar type, undefined and null values are converted to null.
	if v.IsUndefined() || v.IsNull() {
		return js.Null()
	}
	switch scalar {
	case "", nuggit.Bytes, nuggit.String:
		if v.Type() == js.TypeString {
			return v
		}
		// Casting with String yields "[object Object]" which is pointless.
		// stringifying unknown types is much more useful even if its "{}".
		return js.Global().Get("JSON").Call("stringify", v)

	case nuggit.Bool:
		if v.Type() == js.TypeBoolean {
			return v
		}
		if value_isArray(v) {
			// Use an truth assignment that makes empty arrays yield false.
			return js.Global().Call("Boolean", v.Get("length"))
		}
		// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Boolean#boolean_coercion
		return js.Global().Call("Boolean", v) // Coerce boolean.

	case nuggit.Int:
		if v.Type() == js.TypeNumber {
			return v
		}
		// parseInt returns the value, otherwise NaN becomes null.
		res := js.Global().Call("parseInt", v)
		if !res.Truthy() {
			return js.Null()
		}
		return res

	case nuggit.Float:
		if v.Type() == js.TypeNumber && js.Global().Call("isFinite", v).Bool() {
			return v
		}
		// parseFloat returns the value, otherwise NaN becomes null.
		res := js.Global().Call("parseFloat", v)
		if !res.Truthy() {
			return js.Null()
		}
		return res

	default:
		// unexpected scalar type will be JSON stringified
		return js.Global().Get("JSON").Call("stringify", v)
	}
}

// isZero returns whether normalized, casted result value is the zero value with respect to the given point.
//
// isZero returns true for arrays when every value is zero and the point is not repeated.
// If the value is zero, it won't be sent over the exchange.
func value_isZero(v js.Value, scalar nuggit.Scalar) bool {
	if value_isArray(v) {
		// This returns true for empty arrays.
		return v.Call("every", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
			return value_isZeroScalar(args[0], scalar)
		}))).Bool()
	}
	return value_isZeroScalar(v, scalar)
}

func value_isZeroArray(v js.Value, scalar nuggit.Scalar) bool {
	return v.Call("every", js.ValueOf(js.FuncOf(func(_ js.Value, args []js.Value) any {
		return value_isZero(args[0], scalar)
	}))).Bool()
}

func value_isZeroScalar(v js.Value, scalar nuggit.Scalar) bool {
	// Null is always zero.
	if v.IsNull() || v.IsUndefined() {
		return true
	}
	switch scalar {
	case "", nuggit.Bytes, nuggit.String:
		return v.Equal(js.ValueOf("")) // TODO: Use ===

	case nuggit.Bool:
		return v.Equal(js.ValueOf(false)) // TODO: Use ===

	case nuggit.Int, nuggit.Float:
		return v.Equal(js.ValueOf(0)) // TODO: Use ===

	default:
		// unexpected scalar type will be assumed nonzero
		return false
	}
}
