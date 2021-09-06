package lorawan

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/connctd/connector-go"
	"github.com/connctd/lora-connector/lorawan/decoder/ldds75"
	"github.com/connctd/lora-connector/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	connectorClient.On("UpdateThingPropertyValue", mock.MatchedBy(func(in interface{}) bool { return true }), connector.InstantiationToken("abc"), "foothing", "battery", "voltage", "3.336000", mock.AnythingOfType("time.Time")).Return(nil)
	connectorClient.On("UpdateThingPropertyValue", mock.MatchedBy(func(in interface{}) bool { return true }), connector.InstantiationToken("abc"), "foothing", "waterlevel", "waterlevel", "-24.900000", mock.AnythingOfType("time.Time")).Return(nil)

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
