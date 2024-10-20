package api

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/ericchiang/css"
	"github.com/wenooij/nuggit/status"
	"gopkg.in/yaml.v3"
)

type Action struct {
	Action string `json:"action,omitempty"`
	Spec   any    `json:"spec,omitempty"`
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var temp struct {
		Action string          `json:"action,omitempty"`
		Spec   json.RawMessage `json:"spec,omitempty"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal action: %w", err)
	}
	var spec any
	switch temp.Action {
	case ActionAttribute:
		spec = new(AttributeAction)
	case ActionField:
		spec = new(FieldAction)
	case ActionDocument:
		spec = new(DocumentAction)
	case ActionPattern:
		spec = new(PatternAction)
	case ActionSelector:
		spec = new(SelectorAction)
	case ActionPipe:
		spec = new(PipeAction)
	case ActionExchange:
		spec = new(ExchangeAction)
	default:
		return fmt.Errorf("unsupported action (%q): %w", temp.Action, status.ErrInvalidArgument)
	}
	if err := json.Unmarshal(temp.Spec, spec); err != nil {
		return fmt.Errorf("failed to unmarshal spec (%q): %w", temp.Action, err)
	}
	a.Action = temp.Action
	a.Spec = spec
	return nil
}

func (a *Action) UnmarshalYAML(data []byte) error {
	var temp struct {
		Action string    `json:"action,omitempty"`
		Spec   yaml.Node `json:"spec,omitempty"`
	}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal action: %w", err)
	}
	var spec any
	switch temp.Action {
	case ActionAttribute:
		spec = new(AttributeAction)
	case ActionField:
		spec = new(FieldAction)
	case ActionDocument:
		spec = new(DocumentAction)
	case ActionPattern:
		spec = new(PatternAction)
	case ActionSelector:
		spec = new(SelectorAction)
	case ActionPipe:
		spec = new(PipeAction)
	case ActionExchange:
		spec = new(ExchangeAction)
	default:
		return fmt.Errorf("unsupported action (%q): %w", temp.Action, status.ErrInvalidArgument)
	}
	if err := temp.Spec.Decode(spec); err != nil {
		return fmt.Errorf("failed to decode spec (%q): %w", temp.Action, err)
	}
	a.Action = temp.Action
	a.Spec = spec
	return nil
}

func (a *Action) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *Action) GetSpec() any {
	if a == nil {
		return nil
	}
	return a.Spec
}

func (a *Action) GetAttributeAction() *AttributeAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*AttributeAction)
	return spec
}

func (a *Action) GetFieldAction() *FieldAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*FieldAction)
	return spec
}

func (a *Action) GetDocumentAction() *DocumentAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*DocumentAction)
	return spec
}

func (a *Action) GetPatternAction() *PatternAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*PatternAction)
	return spec
}

func (a *Action) GetSelectorAction() *SelectorAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*SelectorAction)
	return spec
}

func (a *Action) GetPipeAction() *PipeAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*PipeAction)
	return spec
}

func (a *Action) GetExchangeAction() *ExchangeAction {
	if a == nil {
		return nil
	}
	spec, _ := a.Spec.(*ExchangeAction)
	return spec
}

// ValidateAction validates the action contents.
//
// Use clientOnly for Pipes, and !clientOnly for Plans.
func ValidateAction(action *Action, clientOnly bool) error {
	if action.GetAction() == "" {
		return fmt.Errorf("action is required: %w", status.ErrInvalidArgument)
	}
	if action.GetSpec() == nil {
		return fmt.Errorf("action spec is required: %w", status.ErrInvalidArgument)
	}
	if clientOnly && action.GetAction() == ActionExchange {
		return fmt.Errorf("exchanges are not allowed in pipes: %w", status.ErrInvalidArgument)
	}
	if _, found := supportedActions[action.GetAction()]; !found {
		return fmt.Errorf("action is not supported (%q): %w", action.GetAction(), status.ErrInvalidArgument)
	}
	spec, ok := action.GetSpec().(interface{ Validate() error })
	if !ok {
		return fmt.Errorf("action spec is not valid (%T): %w", spec, status.ErrInvalidArgument)
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	return nil
}

const (
	ActionAttribute = "attribute" // AttributeAction extracts attribute names from HTML elements.
	ActionField     = "field"     // AttributeAction retrieves fields and or methods from HTML elements.
	ActionDocument  = "document"  // ActionDocument represents an action which copies the full document.
	ActionPattern   = "pattern"   // ActionPattern matches re2 patterns.
	ActionSelector  = "selector"  // ActionSelector matches CSS selectors.
	ActionPipe      = "pipe"      // ActionPipe executes the given pipeline in place.
	ActionExchange  = "exchange"  // ActionExchange submits the result to the server.
)

var supportedActions = map[string]struct{}{
	ActionAttribute: {},
	ActionField:     {},
	ActionDocument:  {},
	ActionPattern:   {},
	ActionSelector:  {},
	ActionPipe:      {},
	ActionExchange:  {},
}

type SelectorAction struct {
	Selector string `json:"selector,omitempty"`
}

func (a *SelectorAction) Validate() error {
	if a.Selector == "" {
		return fmt.Errorf("selector is required: %w", status.ErrInvalidArgument)
	}
	if _, err := css.Parse(a.Selector); err != nil {
		return fmt.Errorf("selector is invalid: %v: %w", err, status.ErrInvalidArgument)
	}
	return nil
}

type DocumentAction struct{}

func (a *DocumentAction) Validate() error {
	*a = DocumentAction{}
	return nil
}

type AttributeAction struct {
	Attribute string `json:"attribute,omitempty"`
}

var attributePattern = regexp.MustCompile(`^(?i:[a-z][a-z0-9-_:.]*)$`)

func (a *AttributeAction) Validate() error {
	if a.Attribute == "" {
		return fmt.Errorf("attribute is required: %w", status.ErrInvalidArgument)
	}
	if !attributePattern.MatchString(a.Attribute) {
		return fmt.Errorf("attribute has invalid characters (%q): %w", a.Attribute, status.ErrInvalidArgument)
	}
	return nil
}

var supportedFields = map[string]struct{}{
	"innerHTML": {},
	"innerText": {},
}

type FieldAction struct {
	Field string `json:"field,omitempty"`
}

func (a *FieldAction) Validate() error {
	if a.Field == "" {
		return fmt.Errorf("field is required: %w", status.ErrInvalidArgument)
	}
	if _, ok := supportedFields[a.Field]; ok {
		return fmt.Errorf("field is not supported (%q): %w", a.Field, status.ErrInvalidArgument)
	}
	return nil
}

type PatternAction struct {
	Pattern string `json:"pattern,omitempty"`
}

func (a *PatternAction) Validate() error {
	if a.Pattern == "" {
		return fmt.Errorf("pattern is required: %w", status.ErrInvalidArgument)
	}
	if _, err := regexp.Compile(a.Pattern); err != nil {
		return fmt.Errorf("not a valid re2 pattern (%q): %v: %w", a.Pattern, err, status.ErrInvalidArgument)
	}
	return nil
}

type PipeAction struct {
	Pipe string `json:"pipe,omitempty"`
}

func (a *PipeAction) Validate() error {
	if a.Pipe == "" {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	// TODO: Clean this up when structured version objects (with optional digest/version) are added.
	if _, err := ParseNameDigest(a.Pipe); err != nil {
		return fmt.Errorf("pipe action has invalid characters (%q): %w", a.Pipe, err)
	}
	return nil
}

type ExchangeAction struct {
	Pipe string `json:"pipe,omitempty"`
}

func (a *ExchangeAction) Validate() error {
	if a.Pipe == "" {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	if _, err := ParseNameDigest(a.Pipe); err != nil {
		return fmt.Errorf("pipe exchanges must reference digest (e.g. pipe@digest; got %q): %w", a.Pipe, err)
	}
	return nil
}
