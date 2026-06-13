package channels

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type channelProtocolExamples struct {
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
}
