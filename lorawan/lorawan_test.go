package lorawan

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/connctd/connector-go"
	_ "github.com/connctd/lora-connector/lorawan/decoder/dcl571"
	"github.com/connctd/lora-connector/lorawan/decoder/ldds75"
	"github.com/connctd/lora-connector/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var dcl571Body = `{"applicationID":"1","applicationName":"AdvantechNode","deviceName":"4610PNode","devEUI":"dP5I//9Edu8=","rxInfo":[],"txInfo":{"frequency":868500000,"modulation":"LORA","loRaModulationInfo":{"bandwidth":125,"spreadingFactor":9,"codeRate":"4/5","polarizationInversion":false}},"adr":false,"dr":3,"fCnt":1316,"fPort":1,"data":"AQwAAKkAgwQBAExhgwQCAOUHgwQDAB0HgwQEAHpFgwQFAON4gwQGAAAAgwQHAAAAgwQIAIBAgwQJAOflgwQKAJlAgwQLAG7OgwQMAI0/gwQNAKH5gwQOAKBCgwQPAAAA","objectJSON":"","tags":{},"confirmedUplink":false,"devAddr":"/0R27w==","publishedAt":"2021-09-15T12:43:37.823217391Z"}`

func TestLDDS75Handling(t *testing.T) {
	connectorClient := new(mocks.Client)
	store := new(mockDataStore)

	loraHandler := NewLoRaWANHandler(connectorClient, true, store)

	store.On("GetInstance", "bar").Return(connector.InstantiationRequest{
		InstallationID: "foo",
		Token:          "abc",
		ID:             "bar",
	}, nil)

	store.On("DecoderNameForApp", "bar", uint64(2)).Return("ldds75", nil)
	store.On("MapDevEUIToThingID", "bar", []byte{0xa8, 0x40, 0x41, 0x4d, 0x61, 0x82, 0xe0, 0x88}).Return("foothing", nil)
	store.On("GetState", "foothing", "mountingHeight").Return([]byte{0xA1}, nil)

	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"battery",
		"voltage",
		"3.336000",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"waterlevel",
		"waterlevel",
		"-24.900000",
		mock.AnythingOfType("time.Time")).Return(nil)

	fr := mux.NewRouter()
	fr.Path("/lora/{installationId}/{instanceId}").Methods(http.MethodPost).Handler(loraHandler)

	buf := &bytes.Buffer{}
	buf.WriteString(ldds75.TestPayload)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/lora/foo/bar?event=up", buf)
	w := httptest.NewRecorder()

	fr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	connectorClient.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestDCL571Handling(t *testing.T) {
	connectorClient := new(mocks.Client)
	store := new(mockDataStore)

	loraHandler := NewLoRaWANHandler(connectorClient, true, store)

	store.On("GetInstance", "bar").Return(connector.InstantiationRequest{
		InstallationID: "foo",
		Token:          "abc",
		ID:             "bar",
	}, nil)

	store.On("DecoderNameForApp", "bar", uint64(1)).Return("dcl571", nil)
	store.On("MapDevEUIToThingID", "bar", []byte{0x74, 0xfe, 0x48, 0xff, 0xff, 0x44, 0x76, 0xef}).Return("foothing", nil)
	store.On("GetState", "foothing", "waterLevelOffset").Return([]byte{0xA1}, nil)

	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"waterlevel",
		"waterlevel",
		"0.402806",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"pressure",
		"pressure",
		"0.000395",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"pressure",
		"maxPressure",
		"4.806449",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"pressure",
		"minPressure",
		"1.109181",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"pressure",
		"pressureUpperLimit",
		"0.392996",
		mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue",
		mock.MatchedBy(func(in interface{}) bool { return true }),
		connector.InstantiationToken("abc"),
		"foothing",
		"pressure",
		"pressureLowerLimit",
		"0.000000",
		mock.AnythingOfType("time.Time")).Return(nil)

	fr := mux.NewRouter()
	fr.Path("/lora/{installationId}/{instanceId}").Methods(http.MethodPost).Handler(loraHandler)

	buf := &bytes.Buffer{}
	buf.WriteString(dcl571Body)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/lora/foo/bar?event=up", buf)
	w := httptest.NewRecorder()

	fr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	connectorClient.AssertExpectations(t)
	store.AssertExpectations(t)

}
