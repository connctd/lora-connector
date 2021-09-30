package restapi

import "time"

// Thing describes a third party device or service
type Thing struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Manufacturer    string           `json:"manufacturer"`
	DisplayType     string           `json:"displayType"`
	MainComponentID string           `json:"mainComponentId"`
	Status          StatusType       `json:"status"`
	Components      []Component      `json:"components"`
	Attributes      []ThingAttribute `json:"attributes,omitempty"`
}

// ThingAttribute defines a key value pair that can be stored within a thing
type ThingAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// StatusType defines thing status
type StatusType string

// Component is part of a thing
type Component struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ComponentType string     `json:"componentType"`
	Capabilities  []string   `json:"capabilities"`
	Properties    []Property `json:"properties,omitempty"`
	Actions       []Action   `json:"actions,omitempty"`
}

// Property is part of a thing component
type Property struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Value        string    `json:"value"`
	Unit         string    `json:"unit"`
	Type         ValueType `json:"type"`
	LastUpdate   time.Time `json:"lastUpdate"`
	PropertyType string    `json:"propertyType"`
}

// ValueType defines the type of a value
type ValueType string

// Action is part of a thing component
type Action struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Parameters []ActionParameter `json:"parameters,omitempty"`
}

// ActionParameter define which parameters a thing action requires
type ActionParameter struct {
	Name string    `json:"name"`
	Type ValueType `json:"type"`
}

// definition of value and status types
const (
	ValueTypeNumber  = ValueType("NUMBER")
	ValueTypeString  = ValueType("STRING")
	ValueTypeBoolean = ValueType("BOOLEAN")

	StatusTypeUnknown     = StatusType("UNKNOWN")
	StatusTypeAvailable   = StatusType("AVAILABLE")
	StatusTypeUnavailable = StatusType("UNAVAILABLE")
)

var (
	// AllValueTypes declares list of possbile value types
	AllValueTypes = map[ValueType]struct{}{
		ValueTypeNumber:  {},
		ValueTypeString:  {},
		ValueTypeBoolean: {},
	}

	// AllStatusTypes defines list of possible status types
	AllStatusTypes = map[StatusType]struct{}{
		StatusTypeUnknown:     {},
		StatusTypeAvailable:   {},
		StatusTypeUnavailable: {},
	}
)
