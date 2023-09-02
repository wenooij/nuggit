// Package jsong implements merging JSON objects.
package jsong

import (
	"reflect"
)

var (
	float64Type = reflect.TypeOf(float64(0))
	boolType    = reflect.TypeOf(false)
	stringType  = reflect.TypeOf("")
)
