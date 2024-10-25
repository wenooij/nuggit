package api

import (
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit/status"
)

type Action map[string]string

func MakeExchangeAction(p *Point, pipe NameDigest) Action {
	a := make(Action, 3)
	a.SetAction("exchange")
	a.SetNameDigest(pipe)
	a["type"] = p.String()
	return a
}

func MakePipeAction(pipe NameDigest) Action {
	a := make(Action, 3)
	a.SetAction("pipe")
	a.SetNameDigest(pipe)
	return a
}

func (a Action) SetNameDigest(nd NameDigest) bool {
	return a.Set("name", nd.Name) && a.Set("digest", nd.Digest)
}

func (a Action) SetAction(action string) bool { return a.Set("action", action) }

func (a Action) Set(key, value string) bool {
	if a == nil {
		return false
	}
	a[key] = value
	return true
}

func (a Action) GetAction() string { return a.GetOrDefaultArg("action") }

func (a Action) GetArg(arg string) (string, bool) {
	v, ok := a[arg]
	return v, ok
}

func (a Action) GetOrDefaultArg(arg string) string {
	v, _ := a[arg]
	return v
}

func (a Action) GetNameDigestArg() NameDigest {
	return NameDigest{
		Name:   a.GetOrDefaultArg("name"),
		Digest: a.GetOrDefaultArg("digest"),
	}
}

func (a Action) writeDigest(h hash.Hash) error {
	return json.NewEncoder(h).Encode(a)
}

// ValidateAction validates the action contents.
//
// Use clientOnly for Pipes, and !clientOnly for Plans.
func ValidateAction(action Action, clientOnly bool) error {
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
