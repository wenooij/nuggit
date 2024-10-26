package api

import (
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
)

func MakeExchangeAction(p nuggit.Point, pipe integrity.NameDigest) nuggit.Action {
	a := make(nuggit.Action, 3)
	a.SetAction("exchange")
	SetActionNameDigest(a, pipe)
	a["type"] = p.String()
	return a
}

func MakePipeAction(pipe integrity.NameDigest) nuggit.Action {
	a := make(nuggit.Action, 3)
	a.SetAction("pipe")
	SetActionNameDigest(a, pipe)
	return a
}

func SetActionNameDigest(a nuggit.Action, nd integrity.NameDigest) bool {
	return a.Set("name", nd.Name) && a.Set("digest", nd.Digest)
}

func writeActionDigest(a nuggit.Action, h hash.Hash) error {
	return json.NewEncoder(h).Encode(a)
}

// ValidateAction validates the action contents.
//
// Use clientOnly for Pipes, and !clientOnly for Plans.
func ValidateAction(action nuggit.Action, clientOnly bool) error {
	if action.GetAction() == "" {
		return fmt.Errorf("action is required: %w", status.ErrInvalidArgument)
	}
	if clientOnly && action.GetAction() == "exchange" {
		return fmt.Errorf("exchanges are not allowed in pipes: %w", status.ErrInvalidArgument)
	}
	if _, found := supportedActions[action.GetAction()]; !found {
		return fmt.Errorf("action is not supported (%q): %w", action.GetAction(), status.ErrInvalidArgument)
	}
	return nil
}

var supportedActions = map[string]struct{}{
	// Nuggit system
	"pipe":     {}, // Execute the specified pipe in place.
	"exchange": {}, // Send the pipe results to the server.

	// Global Objects
	"regexp": {}, // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Map/get
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/at
	"index": {},
	"split": {}, // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/split

	// Document
	"documentElement": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Document/documentElement

	// HTML Elements
	"innerHTML":       {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/innerHTML
	"innerText":       {}, // https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/innerText
	"attributes":      {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/attributes
	"querySelector":   {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelector
	"querSelectorAll": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelectorAll
}
