package properties

import (
	"encoding/json"
	"testing"

	"github.com/fu2hito/go-liveview/internal/protocol"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestMessageEncodeDecodeRoundTrip tests that encoding then decoding returns the original message
func TestMessageEncodeDecodeRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("decode(encode(msg)) == msg", prop.ForAll(
		func(topic, event string, payload map[string]string) bool {
			// Convert to interface map
			payloadInterface := make(map[string]interface{})
			for k, v := range payload {
				payloadInterface[k] = v
			}

			payloadBytes, _ := json.Marshal(payloadInterface)
			original := &protocol.Message{
				Topic:   topic,
				Event:   event,
				Payload: payloadBytes,
			}

			encoded, err := original.Encode()
			if err != nil {
				return false
			}

			decoded, err := protocol.DecodeMessage(encoded)
			if err != nil {
				return false
			}

			return decoded.Topic == original.Topic &&
				decoded.Event == original.Event &&
				string(decoded.Payload) == string(original.Payload)
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	))

	properties.TestingRun(t)
}

// TestEventPayloadEncodeDecode tests EventPayload round-trip
func TestEventPayloadEncodeDecode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()

	properties := gopter.NewProperties(parameters)

	properties.Property("EventPayload round-trip", prop.ForAll(
		func(eventType, eventName, target string, value map[string]string) bool {
			valueInterface := make(map[string]interface{})
			for k, v := range value {
				valueInterface[k] = v
			}

			original := protocol.EventPayload{
				Type:   eventType,
				Event:  eventName,
				Value:  valueInterface,
				Target: target,
			}

			encoded, err := json.Marshal(original)
			if err != nil {
				return false
			}

			var decoded protocol.EventPayload
			err = json.Unmarshal(encoded, &decoded)
			if err != nil {
				return false
			}

			return decoded.Type == original.Type &&
				decoded.Event == original.Event &&
				decoded.Target == original.Target
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	))

	properties.TestingRun(t)
}
