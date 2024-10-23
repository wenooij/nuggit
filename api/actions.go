package api

import (
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit/status"
)

type Action struct {
	Action string            `json:"action,omitempty"`
	Args   map[string]string `json:"args,omitempty"`
}

func (a *Action) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *Action) GetArgs() map[string]string {
	if a == nil {
		return nil
	}
	return a.Args
}

func (a *Action) GetArg(arg string) (string, bool) {
	v, ok := a.GetArgs()[arg]
	return v, ok
}

func (a *Action) GetArgDefault(arg string) string {
	v, _ := a.GetArgs()[arg]
	return v
}

func (a *Action) GetPipeArg() NameDigest {
	switch a.GetAction() {
	case "pipe":
		return NameDigest{
			Name:   a.GetArgDefault("name"),
			Digest: a.GetArgDefault("digest"),
		}

	default:
		return NameDigest{}
	}
}

func (a *Action) writeDigest(h hash.Hash) error {
	return json.NewEncoder(h).Encode(a)
}

// ValidateAction validates the action contents.
//
// Use clientOnly for Pipes, and !clientOnly for Plans.
func ValidateAction(action *Action, clientOnly bool) error {
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
	"pipe":     {}, // ActionPipe executes the given pipeline in place.
	"exchange": {}, // ActionExchange submits the result to the server.

	// Global Objects
	"regexp": {}, // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Map/get
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/at
	"index": {},

	// Document
	"documentElement": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Document/documentElement

	// HTML Elements
	"innerHTML":       {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/innerHTML
	"innerText":       {}, // https://developer.mozilla.org/en-US/docs/Web/API/HTMLElement/innerText
	"attributes":      {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/attributes
	"querySelector":   {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelector
	"querSelectorAll": {}, // https://developer.mozilla.org/en-US/docs/Web/API/Element/querySelectorAll
}
