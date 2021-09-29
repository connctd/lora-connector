package dcl571

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var base64Payload = `AQwAAKkAgwQBAExhgwQCAOUHgwQDAB0HgwQEAHpFgwQFAON4gwQGAAAAgwQHAAAAgwQIAIBAgwQJAOflgwQKAJlAgwQLAG7OgwQMAI0/gwQNAKH5gwQOAKBCgwQPAAAA`

func TestDecodePressureValue(t *testing.T) {
	payload, err := base64.RawStdEncoding.DecodeString(base64Payload)
	require.NoError(t, err)

	pressure := decodePressureValue(payload[52:])
	assert.EqualValues(t, float32(4.0280643), pressure)

	maxPressure := decodePressureValue(payload[28:])
	assert.EqualValues(t, float32(4007.555420), maxPressure)
}
