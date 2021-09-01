package decoder

import (
	"fmt"
	"time"

	"github.com/connctd/restapi-go"
)

var decoders = map[string]PayloadDecoder{}

func RegisterDecoder(name string, decoder PayloadDecoder) {
	if _, exists := decoders[name]; exists {
		panic(fmt.Errorf("PayloadDecoder with name %s alrady exists", name))
	}
	decoders[name] = decoder
}

func GetDecoder(name string) PayloadDecoder {
	return decoders[name]
}

type PropertyUpdate struct {
	ThingID     string
	ComponentID string
	PropertyID  string
	Value       string
	UpdateTime  time.Time
}

type PayloadDecoder interface {
	Device(attributes []restapi.ThingAttribute) (*restapi.Thing, error)
	DecodeMessage(fport uint32, msg []byte, thingID string) ([]PropertyUpdate, error)
}
