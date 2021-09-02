package lorawan

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/connctd/connector-go"
	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
)

type dataStore interface {
	MapDevEUIToThingID(instanceID string, devEUI []byte) (string, error)
	StoreDEVUIToThingID(instanceID string, devEUI []byte, thingID string) error
	GetInstallationToken(installationId string) (connector.InstallationToken, error)
	GetInstance(instanceId string) (connector.InstantiationRequest, error)
	DecoderNameForApp(instanceID string, appId uint64) (string, error)
}
type LoRaWANHandler struct {
	json            bool
	connectorClient connector.Client
	logger          logrus.FieldLogger
	store           dataStore
}

func NewLoRaWANHandler(connectorClient connector.Client, useJson bool, dataStore dataStore) *LoRaWANHandler {
	l := &LoRaWANHandler{
		connectorClient: connectorClient,
		store:           dataStore,
		json:            useJson,
		logger:          logrus.WithField("component", "LoRaWANHandler"),
	}

	return l
}

// Handler under form of .../{installationId}/{instanceId}
func (l *LoRaWANHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	installationID := vars["installationId"]

	if installationID == "" || instanceID == "" {
		l.logger.WithField("url", r.RequestURI).Error("Invalid request URL. Either instanceId or installationId are missing")
		http.Error(w, "invalid request url", http.StatusBadRequest)
		return
	}
	logger := l.logger.WithFields(logrus.Fields{
		"instanceId":     instanceID,
		"installationId": installationID,
	})
	instance, err := l.store.GetInstance(installationID)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve installation token")
		http.Error(w, "internal failure", http.StatusInternalServerError)
		return
	}
	if installationID != instance.InstallationID {
		logger.WithFields(logrus.Fields{
			"storedInstallationId": instance.InstallationID,
		}).Error("The installation id received and stored do not match")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	l.HandleRequest(instance.Token, installationID, w, r)
}

func (l *LoRaWANHandler) HandleRequest(token connector.InstantiationToken, instanceID string, w http.ResponseWriter, r *http.Request) {
	event := r.URL.Query().Get("event")
	logger := l.logger.WithField("event", event)

	b, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024*1024*1)) // Read up to 1 Megabyte, the calls shouldn't be bigger than this
	if err != nil {
		logger.WithError(err).Error("Failed to read body of callback request")
		http.Error(w, "failed to read request", http.StatusBadRequest)
		return
	}

	switch event {
	case "up":
		//err = h.up(b)
		var up integration.UplinkEvent
		err = l.unmarshal(b, &up)
		if err != nil {
			logger.WithError(err).Error("Failed to unmarshal http callback payload")
			http.Error(w, "unparseable payload", http.StatusBadRequest)
			return
		}
		// applicationId := formatAppID(up.ApplicationId)
		decoderName, err := l.store.DecoderNameForApp(instanceID, up.ApplicationId)
		if err != nil {
			logger.WithError(err).Error("Failed to retrieve decoder name for LoRaWAN application")
			return
		}
		payloadDecoder := decoder.GetDecoder(decoderName)
		if payloadDecoder == nil {
			logger.WithField("decoderName", decoderName).Error("For this name no payload decoder implementation is registered")
			return
		}
		deviceID := up.DevEui
		formattedEUI, err := formatEUI(deviceID)
		if err != nil {
			l.logger.WithError(err).Error("Failed to format DeviceEUI")
			formattedEUI = "invalid EUI"
		}
		msg := up.Data
		fport := up.FPort
		logger = logger.WithFields(logrus.Fields{
			"deviceID": formattedEUI,
			"fport":    fport,
		})

		thingID, err := l.store.MapDevEUIToThingID(instanceID, deviceID)
		if err != nil {
			logger.WithError(err).Error("Failed to retrieve thingID for device EUI")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if thingID == "" {
			attributes := []restapi.ThingAttribute{
				{
					Name:  "lora.deveui",
					Value: formattedEUI,
				},
			}
			thing, err := payloadDecoder.Device(attributes)
			if err != nil {
				logger.WithError(err).Error("Failed to create thing for LoRaWAN device")
				http.Error(w, "unable to create thing", http.StatusInternalServerError)
				return
			}
			result, err := l.connectorClient.CreateThing(r.Context(), token, *thing)
			if err != nil {
				logger.WithError(err).Error("Failed to create thing")
				http.Error(w, "upstream error", http.StatusInternalServerError)
				return
			}
			if err := l.store.StoreDEVUIToThingID(instanceID, deviceID, result.ID); err != nil {
				logger.WithError(err).Error("Failed to store deviceEUI to thing ID mapping")
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			thingID = result.ID
		}

		logger = logger.WithField("thingID", thingID)

		updates, err := payloadDecoder.DecodeMessage(fport, msg, thingID)
		if err != nil {
			logger.WithField("deviceEUI", deviceID).WithError(err).Error("Failed to decode message of LoRaWAN device")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		for _, update := range updates {
			updateTime := update.UpdateTime
			if updateTime.IsZero() {
				updateTime = time.Now()
			}
			if err := l.connectorClient.UpdateThingPropertyValue(
				r.Context(),
				token,
				update.ThingID,
				update.ComponentID,
				update.PropertyID,
				update.Value,
				updateTime); err != nil {
				logger.WithFields(logrus.Fields{
					"componentId": update.ComponentID,
					"propertyId":  update.PropertyID,
				}).WithError(err).Error("Failed to update thing property")
			}
		}

	case "error":
		var errorEvent integration.ErrorEvent
		if err := l.unmarshal(b, &errorEvent); err != nil {
			logger.WithError(err).Error("Failed to unmarshal error event")
			http.Error(w, "unparseable payload", http.StatusBadRequest)
			return
		}
		formattedEUI, err := formatEUI(errorEvent.DevEui)
		if err != nil {
			logger.WithError(err).Error("Failed to format device EUI")
			formattedEUI = "invalid Device EUI"
		}
		logger.WithFields(logrus.Fields{
			"deviceEUI":       formattedEUI,
			"errorType":       errorEvent.Type.String(),
			"applicationName": errorEvent.ApplicationName,
			"applicationId":   errorEvent.ApplicationId,
		}).Error(errorEvent.Error)
	default:
		logger.Warn("Handler for event type not implemented")
		http.Error(w, "unknown event type", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("handling event '%s' returned error: %s", event, err)
	}
}

func (l *LoRaWANHandler) unmarshal(b []byte, v proto.Message) error {
	// TODO figure out automaticall (via headers or something) if we should decode json or protobuf
	if l.json {
		unmarshaler := &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		}
		return unmarshaler.Unmarshal(bytes.NewReader(b), v)
	}
	return proto.Unmarshal(b, v)
}

func formatEUI(eui []byte) (string, error) {
	if len(eui) != 8 {
		return "", errors.New("invalid EUI. Invalid length")
	}
	return fmt.Sprintf("%#X-%#X-%#X-%#X-%#X-%#X-%#X-%#X", eui[0], eui[1], eui[2], eui[3], eui[4], eui[5], eui[6], eui[7]), nil
}

func formatAppID(appID uint64) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, appID)
	s, _ := formatEUI(b)
	return s
}
