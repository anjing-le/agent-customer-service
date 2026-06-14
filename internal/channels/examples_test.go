package channels

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type channelProtocolExamples struct {
	PlatformSignatureProfiles []struct {
		ID                     string   `json:"id"`
		DemoSecret             string   `json:"demoSecret"`
		SampleCanonicalPayload []string `json:"sampleCanonicalPayload"`
		SampleSignature        string   `json:"sampleSignature"`
	} `json:"platformSignatureProfiles"`
	Examples []struct {
		ID             string `json:"id"`
		DemoSecret     string `json:"demoSecret"`
		SignatureInput struct {
			Channel                string `json:"channel"`
			ExternalConversationID string `json:"externalConversationId"`
			Timestamp              string `json:"timestamp"`
			Content                string `json:"content"`
		} `json:"signatureInput"`
		Request struct {
			Signature string `json:"signature"`
		} `json:"request"`
	} `json:"examples"`
}

func TestChannelProtocolExamplesMatchSignatureImplementation(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "contracts", "examples", "channel-protocols.json"))
	if err != nil {
		t.Fatalf("read channel protocol examples: %v", err)
	}
	var examples channelProtocolExamples
	if err := json.Unmarshal(raw, &examples); err != nil {
		t.Fatalf("decode channel protocol examples: %v", err)
	}
	if len(examples.Examples) == 0 {
		t.Fatal("expected channel protocol examples")
	}

	for _, item := range examples.Examples {
		input := item.SignatureInput
		expected := ChannelSignatureWithSecret(item.DemoSecret, input.Channel, input.ExternalConversationID, input.Timestamp, input.Content)
		if item.Request.Signature != expected {
			t.Fatalf("%s signature mismatch: expected %s, got %s", item.ID, expected, item.Request.Signature)
		}
	}
	for _, profile := range examples.PlatformSignatureProfiles {
		if len(profile.SampleCanonicalPayload) != 4 {
			t.Fatalf("%s expected four canonical fields, got %d", profile.ID, len(profile.SampleCanonicalPayload))
		}
		expected := ChannelSignatureWithSecret(
			profile.DemoSecret,
			profile.SampleCanonicalPayload[0],
			profile.SampleCanonicalPayload[1],
			profile.SampleCanonicalPayload[2],
			profile.SampleCanonicalPayload[3],
		)
		if profile.SampleSignature != expected {
			t.Fatalf("%s signature mismatch: expected %s, got %s", profile.ID, expected, profile.SampleSignature)
		}
	}
}
