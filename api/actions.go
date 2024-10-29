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
	setActionNameDigest(a, pipe)
	a["type"] = p.String()
	return a
}

func MakePipeAction(pipe integrity.NameDigest) nuggit.Action {
	a := make(nuggit.Action, 3)
	a.SetAction("pipe")
	setActionNameDigest(a, pipe)
	return a
}

func setActionNameDigest(a nuggit.Action, nd integrity.NameDigest) {
	a.Set("name", nd.GetName())
	a.Set("digest", nd.GetDigest())
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
	"get":    {}, // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/get#prop
	"split":  {}, // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/split

	// Document
	"documentElement": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Document/documentElement

	// HTML Elements
	"innerHTML":        {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/innerHTML
	"outerHTML":        {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/outerHTML
	"innerText":        {}, // https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/innerText
	"attributes":       {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/attributes
	"querySelector":    {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelector
	"querySelectorAll": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelectorAll
}
