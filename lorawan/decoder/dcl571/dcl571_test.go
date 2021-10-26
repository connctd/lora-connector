package dcl571

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var base64Payloads = []string{
	`gc1PAAcHAgAAAAAAAAcnAgAAAAAANAofIACAAAAVgNx/NAo/IACAAAAGgPl/EAIBAIMEAACpAIMEAQBMYYMEAgADQ4MEAwDpGGAHFgJigKV2YTU`,
	`gd1PAAcHAgAAAAAAAAcnAgAAAAAANAofIACAAAAVgNx/NAo/IP9/AAAGgPl/EAIBAIMEAACpAIMEAQBMYYMEAgADQ4MEAwCLfGAHFgJhq8t2YaQ`,
}

func TestDecodePressureValue(t *testing.T) {
	payload, err := base64.RawStdEncoding.DecodeString(base64Payloads[0])
	require.NoError(t, err)

	pressure, _ := decodePressureValue(payload[65:])
	assert.EqualValues(t, float32(131.0973), pressure)

	payload, err = base64.RawStdEncoding.DecodeString(base64Payloads[1])
	require.NoError(t, err)

	pressure, _ = decodePressureValue(payload[65:])
	assert.EqualValues(t, float32(131.486496), pressure)
}
