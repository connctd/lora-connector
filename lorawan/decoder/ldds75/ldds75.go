package ldds75

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
	"gorm.io/gorm"
)

func init() {
	decoder.RegisterDecoder("ldds75", ldds75decoder{})
}

var TestPayload = `{"applicationID":"2","applicationName":"newapp","deviceName":"ldds75","devEUI":"qEBBTWGC4Ig=","rxInfo":[{"gatewayID":"dP5I//5MnTc=","time":"2021-08-15T12:06:41.391514Z","timeSinceGPSEpoch":null,"rssi":-56,"loRaSNR":9.5,"channel":3,"rfChain":0,"board":0,"antenna":0,"location":{"latitude":0,"longitude":0,"altitude":0,"source":"UNKNOWN","accuracy":0},"fineTimestampType":"NONE","context":"XywUZA==","uplinkID":"ZRepk5N3QNWwfHYQQeBenQ==","crcStatus":"CRC_OK"}],"txInfo":{"frequency":867100000,"modulation":"LORA","loRaModulationInfo":{"bandwidth":125,"spreadingFactor":12,"codeRate":"4/5","polarizationInversion":false}},"adr":true,"dr":0,"fCnt":33,"fPort":2,"data":"DQgA+QA=","objectJSON":"{\"Bat\":3.336,\"Distance\":249,\"Interrupt_status\":0}","tags":{},"confirmedUplink":false,"devAddr":"AFOWqg==","publishedAt":"2021-08-15T12:06:41.521032244Z"}`

type ldds75decoder struct{}

func (d ldds75decoder) Device(attributes []restapi.ThingAttribute) (*restapi.Thing, error) {
	return &restapi.Thing{
		Name:            "LDDS75",
		Manufacturer:    "Dragino",
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
				ID:            "configuration",
				Name:          "Configuration",
				ComponentType: "dragino.CONFIGURATION",
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
						ID:   "setMountingHeight",
						Name: "SetMountingHeight",
						Parameters: []restapi.ActionParameter{
							{
								Name: "mountingHeight",
								Type: restapi.ValueTypeNumber,
							},
						},
					},
				},
			},
			{
				ID:            "battery",
				Name:          "Battery",
				ComponentType: "core.BATTERY",
				Capabilities:  []string{"core.MEASURE"},
				Properties: []restapi.Property{
					{
						ID:           "voltage",
						Name:         "Voltage",
						Value:        "",
						Unit:         "VOLT",
						PropertyType: "core.VOLTAGE",
					}, {
						ID:           "chemistry",
						Name:         "Battery chemistry",
						Value:        "Li-SoCl2",
						Unit:         "",
						PropertyType: "core.STRING",
					},
				},
			},
		},
	}, nil
}

func (d ldds75decoder) DecodeMessage(store decoder.DecoderStateStore, fport uint32, msg []byte, thingID string) ([]decoder.PropertyUpdate, error) {
	// Ignore fport, device seems to only transmit on port 2
	if len(msg) < 2 {
		return nil, errors.New("message shorter than 2 bytes")
	}
	updates := []decoder.PropertyUpdate{}
	var battInt uint16
	battInt = binary.BigEndian.Uint16(msg[:2]) //

	battInt = battInt & 0x3FFF
	battVoltage := float32(battInt) / 1000.0 // battery voltage in V
	updates = append(updates, decoder.PropertyUpdate{
		ThingID:     thingID,
		ComponentID: "battery",
		PropertyID:  "voltage",
		Value:       fmt.Sprintf("%f", battVoltage),
		UpdateTime:  time.Now(),
	})
	// HINT calculating battery percentage is not very useful since Li-SoCl2 batteries seem to have a
	// pretty narrow voltage difference between full and empty

	if len(msg) >= 5 {
		// if message is shorter, the distance sensor is not connected. We should signal this somehow
		distanceRaw := binary.BigEndian.Uint16(msg[2:4]) // distance in mm
		if distanceRaw > 20 {                            // Values smaller than 20 indicate invalid readings

			val, err := store.GetState(thingID, "mountingHeight")
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					val = make([]byte, 4)
					n := binary.PutVarint(val, 0)
					err = store.SetState(thingID, "mountingHeight", val[:n])
					if err != nil {
						return nil, err
					}
					val = val[:n]
				} else {
					return nil, err
				}

			}
			mountingHeightInt, _ := binary.Varint(val)
			mountingHeight := float64(mountingHeightInt) / 10.0
			updates = append(updates, decoder.PropertyUpdate{
				ThingID:     thingID,
				ComponentID: "waterlevel",
				PropertyID:  "waterlevel",
				Value:       fmt.Sprintf("%f", (mountingHeight - float64(distanceRaw)/10.0)),
			})

		}
	}
	// We ignore the interrupt for now

	return updates, nil
}
