package dcl571

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
	"github.com/sirupsen/logrus"
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
	logger := logrus.WithFields(logrus.Fields{
		"thingId": thingID,
		"fport":   fport,
		"decoder": "dcl571",
		"msgLen":  len(msg),
	})

	if len(msg) >= 65+7 {
		pressure, err := decodePressureValue(msg[65:])
		if err != nil {
			logger.WithError(err).WithField("msg", hex.EncodeToString(msg)).Warn("Failed to decode pressure")
		} else {
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

	}

	return updates, nil
}

func decodePressureValue(msg []byte) (float32, error) {
	if len(msg) < 8 {
		return 0.0, errors.New("message to small to contain valid pressure value")
	}
	b := []byte{msg[6], msg[7], msg[0], msg[1]}
	buf := bytes.NewReader(b)
	var val float32
	binary.Read(buf, binary.LittleEndian, &val)
	return val, nil
}
