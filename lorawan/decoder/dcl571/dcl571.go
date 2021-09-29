package dcl571

import (
	"bytes"
	"encoding/binary"

	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
)

func init() {
	decoder.RegisterDecoder("dcl571", dcl571decoder{})
}

type dcl571decoder struct{}

func (d dcl571decoder) Device(attributes []restapi.ThingAttribute) (*restapi.Thing, error) {
	return &restapi.Thing{
		Name:            "DCL571",
		Manufacturer:    "Bode",
		DisplayType:     "SENSOR",
		Status:          restapi.StatusTypeAvailable,
		Attributes:      attributes,
		MainComponentID: "waterlevel",
		Components: []restapi.Component{
			{
				ID:            "waterlevel",
				Name:          "Messstelle",
				ComponentType: "core.SENSOR",
				Capabilities:  []string{"core.MEASURE"},
				Properties: []restapi.Property{
					{
						ID:           "waterlevel",
						Name:         "Waterlevel",
						Value:        "",
						Unit:         "CENTIMETER",
						PropertyType: "core.NUMBER",
					},
				},
			},
			{
				ID:            "pressure",
				Name:          "Druck",
				ComponentType: "core.SENSOR",
				Capabilities:  []string{"core.MEASURE"},
				Properties: []restapi.Property{
					{
						ID:           "pressure",
						Name:         "pressure",
						Value:        "",
						Unit:         "Bar",
						PropertyType: "core.NUMBER",
					},
				},
			},
			{
				ID:            "configuration",
				Name:          "Configuration",
				ComponentType: "bode.CONFIGURATION",
				Capabilities:  []string{},
				Properties: []restapi.Property{
					{
						ID:           "mountingHeight",
						Name:         "Mounting Height",
						Value:        "0",
						Unit:         "CENTIMETER",
						PropertyType: "core.NUMBER",
					},
				},
				Actions: []restapi.Action{
					{
						ID:   "setWaterLevelOffset",
						Name: "setWaterLevelOffset",
						Parameters: []restapi.ActionParameter{
							{
								Name: "offset",
								Type: restapi.ValueTypeNumber,
							},
						},
					},
				},
			},
		},
	}, nil
}

func (d dcl571decoder) DecodeMessage(store decoder.DecoderStateStore, fport uint32, msg []byte, thingID string) ([]decoder.PropertyUpdate, error) {
	updates := []decoder.PropertyUpdate{}

	// pressureBytes := msg[53:59] //54:60

	return updates, nil
}

func decodePressureValue(msg []byte) float32 {
	b := []byte{msg[6], msg[7], msg[0], msg[1]}
	buf := bytes.NewReader(b)
	var val float32
	binary.Read(buf, binary.LittleEndian, &val)
	return val
}
