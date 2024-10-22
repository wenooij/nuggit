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
	if temp.Spec == nil {
		temp.Spec = []byte("null")
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal action: %w", err)
	}
	spec, err := NewActionSpec(temp.Action)
	if err != nil {
		return err
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
	spec, err := NewActionSpec(temp.Action)
	if err != nil {
		return err
	}
	if err := temp.Spec.Decode(spec); err != nil {
		return fmt.Errorf("failed to decode spec (%q): %w", temp.Action, err)
	}
	a.Action = temp.Action
	a.Spec = spec
	return nil
}

func (a *Action) Equal(b *Action) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Action != b.Action {
		return false
	}
	switch a.Action {
	case ActionAttribute:
		return equalSpec[*AttributeAction](a.Spec, b.Spec)
	case ActionField:
		return equalSpec[*FieldAction](a.Spec, b.Spec)
	case ActionDocument:
		return equalSpec[*DocumentAction](a.Spec, b.Spec)
	case ActionPattern:
		return equalSpec[*PatternAction](a.Spec, b.Spec)
	case ActionSelector:
		return equalSpec[*SelectorAction](a.Spec, b.Spec)
	case ActionPipe:
		return equalSpec[*PipeAction](a.Spec, b.Spec)
	case ActionExchange:
		return equalSpec[*ExchangeAction](a.Spec, b.Spec)
	default:
		return false
	}
}

func equalSpec[T *E, E comparable](spec, spec2 any) bool {
	a, ok := spec.(T)
	if !ok {
		return false
	}
	b, ok := spec2.(T)
	if !ok {
		return false
	}
	// Equate nil and zero.
	var zero E
	return a == b || (a == nil && *b == zero) || (b == nil && *a == zero) || *a == *b
}

func (a *Action) GetAction() string {
	if a == nil {
		return ""
	}
	return a.Action
}

func (a *Action) GetName() string { return a.GetAction() }

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

func NewActionSpec(action string) (any, error) {
	switch action {
	case ActionAttribute:
		return new(AttributeAction), nil
	case ActionField:
		return new(FieldAction), nil
	case ActionDocument:
		return new(DocumentAction), nil
	case ActionPattern:
		return new(PatternAction), nil
	case ActionSelector:
		return new(SelectorAction), nil
	case ActionPipe:
		return new(PipeAction), nil
	case ActionExchange:
		return new(ExchangeAction), nil
	default:
		return nil, fmt.Errorf("unsupported action (%q): %w", action, status.ErrInvalidArgument)
	}
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
	if _, ok := supportedFields[a.Field]; !ok {
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
	Pipe NameDigest `json:"pipe,omitempty"`
}

func (a *PipeAction) Validate() error {
	if a.Pipe.GetName() == "" {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	if err := ValidateNameDigest(a.Pipe); err != nil {
		return fmt.Errorf("pipe action is invalid: %w", err)
	}
	return nil
}

type ExchangeAction struct {
	Pipe NameDigest `json:"pipe,omitempty"`
}

func (a *ExchangeAction) Validate() error {
	if a.Pipe.GetName() == "" {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	if err := ValidateNameDigest(a.Pipe); err != nil {
		return fmt.Errorf("exhcnage action is invalid (does it reference a pipe@digest?): %w", err)
	}
	return nil
}
