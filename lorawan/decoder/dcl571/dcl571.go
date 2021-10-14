package dcl571

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
	"gorm.io/gorm"
)

const metersPerBar = 10.1972

const mmH2OPerBar = 0.0000980638

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
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
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
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
					},
					{
						ID:           "maxPressure",
						Name:         "maxPressure",
						Value:        "",
						Unit:         "Bar",
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
					},
					{
						ID:           "minPressure",
						Name:         "minPressure",
						Value:        "",
						Unit:         "Bar",
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
					},
					{
						ID:           "pressureUpperLimit",
						Name:         "pressureUpperLimit",
						Value:        "",
						Unit:         "Bar",
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
					},
					{
						ID:           "pressureLowerLimit",
						Name:         "pressureLowerLimit",
						Value:        "",
						Unit:         "Bar",
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
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
						ID:           "waterLevelOffset",
						Name:         "water level offset",
						Value:        "0",
						Unit:         "CENTIMETER",
						PropertyType: "NUMBER",
						Type:         restapi.ValueTypeNumber,
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

	if len(msg) > 28+7 {
		upperPressureLimit := decodePressureValue(msg[28:])
		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "pressure",
			PropertyID:  "pressureUpperLimit",
			Value:       fmt.Sprintf("%f", upperPressureLimit*mmH2OPerBar),
			UpdateTime:  time.Now(),
		})
	}

	if len(msg) > 40+7 {
		lowerPressureLimit := decodePressureValue(msg[40:])
		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "pressure",
			PropertyID:  "pressureLowerLimit",
			Value:       fmt.Sprintf("%f", lowerPressureLimit*mmH2OPerBar),
			UpdateTime:  time.Now(),
		})
	}

	if len(msg) > 52+7 {
		pressure := decodePressureValue(msg[52:])
		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "pressure",
			PropertyID:  "pressure",
			Value:       fmt.Sprintf("%f", pressure*mmH2OPerBar),
			UpdateTime:  time.Now(),
		})

		val, err := store.GetState(thingID, "waterLevelOffset")
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				val = make([]byte, 4)
				n := binary.PutVarint(val, 0)
				err = store.SetState(thingID, "waterLevelOffset", val[:n])
				if err != nil {
					return nil, err
				}
				val = val[:n]
			} else {
				return nil, err
			}
		}
		waterLevelOffsetMM, _ := binary.Varint(val)
		waterLevel := (float32(waterLevelOffsetMM) + pressure) / 10.0

		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "waterlevel",
			PropertyID:  "waterlevel",
			Value:       fmt.Sprintf("%f", waterLevel),
			UpdateTime:  time.Now(),
		})
	}

	if len(msg) > 64+1 {
		maxPressure := decodePressureValue(msg[64:])
		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "pressure",
			PropertyID:  "maxPressure",
			Value:       fmt.Sprintf("%f", maxPressure),
			UpdateTime:  time.Now(),
		})
	}

	if len(msg) > 76+1 {
		minPressure := decodePressureValue(msg[76:])
		updates = append(updates, decoder.PropertyUpdate{
			ThingID:     thingID,
			ComponentID: "pressure",
			PropertyID:  "minPressure",
			Value:       fmt.Sprintf("%f", minPressure),
			UpdateTime:  time.Now(),
		})
	}

	return updates, nil
}

func decodePressureValue(msg []byte) float32 {
	b := []byte{msg[6], msg[7], msg[0], msg[1]}
	buf := bytes.NewReader(b)
	var val float32
	binary.Read(buf, binary.LittleEndian, &val)
	return val
}
