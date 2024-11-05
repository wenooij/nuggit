package client

// Envelope wraps all messages between the browser and the service worker.
//
// We do this to mux messages by a type name.
type Envelope struct {
	Type string `json:"type,omitempty"`
	Data any    `json:",omitempty"`
}

// Navigate contains the page's URL.
//
// Navigate is sent from the browser to the service worker when the page's URL has changes.
type Navigate struct {
	URL string `json:"url,omitempty"`
}

// Observe contains a list of rules to observe.
//
// Observe is sent from the service worker to the browser with a list of rules to observe.
type Observe struct {
	Rules []Rule `json:"filter_list,omitempty"`
}

// Results contains the processed elements as the result of observing DOM changes with the given filters.
type Results struct {
	Elements []Element
}

// Rule contains a filter and a list of Actions.
//
// Rule is sent as part of the Observe message.
type Rule struct {
	Filter Filter `json:"filter,omitempty"`
	Action Action `json:"action,omitempty"`
}

// Filter contains a CSS selector which matches Elements on the page.
//
// Filter is sent as part of the Rule message.
type Filter struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	NodeType       string `json:"node_type,omitempty"`
	Class          string `json:"class,omitempty"`
	Attribute      string `json:"attribute,omitempty"`
	AttributeValue string `json:"attribute_value,omitempty"`
	AttributeEmpty bool   `json:"attribute_empty,omitempty"`
	Selector       string `json:"selector,omitempty"`
}

// Action describes which Element fields should be populated for matched Elements.
type Action struct {
	ID              bool     `json:"id,omitempty"`
	Name            bool     `json:"name,omitempty"`
	NodeType        bool     `json:"node_type,omitempty"`
	Class           bool     `json:"class,omitempty"`
	Attributes      []string `json:"attributes,omitempty"`
	AttributeValues bool     `json:"attribute_values,omitempty"`
	InnerText       bool     `json:"inner_text,omitempty"`
	InnerHTML       bool     `json:"inner_html,omitempty"`
	OuterHTML       bool     `json:"outer_html,omitempty"`
	TextContent     bool     `json:"text_content,omitempty"`
	// For <canvas> elements.
	GraphicsContext string `json:"graphics_context,omitempty"`
	PixelData       bool   `json:"pixel_data,omitempty"`
}

// Element contains a serializable version of the observable parts of a JS element.
//
// Not all fields need be set for a given instance.
type Element struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	NodeType    string            `json:"node_type,omitempty"`
	Class       string            `json:"class,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	InnerText   string            `json:"inner_text,omitempty"`
	InnerHTML   string            `json:"inner_html,omitempty"`
	OuterHTML   string            `json:"outer_html,omitempty"`
	TextContent string            `json:"text_content,omitempty"`
	// For <canvas> elements.
	GraphicsContext string  `json:"graphics_context,omitempty"`
	PixelData       []uint8 `json:"pixel_data,omitempty"`
}
