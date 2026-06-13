package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anjing-le/agent-customer-service/internal/platform/store"
)

type Client struct {
	apiURL     string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiURL, apiKey, model string) *Client {
	return &Client{
		apiURL: strings.TrimSpace(apiURL),
		apiKey: strings.TrimSpace(apiKey),
		model:  fallback(strings.TrimSpace(model), "gpt-4o-mini"),
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiURL != "" && c.apiKey != ""
}

func (c *Client) GenerateReply(ctx context.Context, req store.ReplyRequest) (store.ReplyGeneration, error) {
	if !c.Enabled() {
		return store.ReplyGeneration{}, fmt.Errorf("llm client is not configured")
	}
	if len(req.Evidence) == 0 {
		return store.ReplyGeneration{}, fmt.Errorf("evidence is required")
	}

	body := chatRequest{
		Model:       c.model,
		Temperature: 0.2,
		Messages:    c.messages(req),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return store.ReplyGeneration{}, fmt.Errorf("marshal llm request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(payload))
	if err != nil {
		return store.ReplyGeneration{}, fmt.Errorf("create llm request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return store.ReplyGeneration{}, fmt.Errorf("call llm: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return store.ReplyGeneration{}, fmt.Errorf("llm status %d", resp.StatusCode)
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return store.ReplyGeneration{}, fmt.Errorf("decode llm response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return store.ReplyGeneration{}, fmt.Errorf("llm response has no choices")
	}
	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return store.ReplyGeneration{}, fmt.Errorf("llm response is empty")
	}
	return store.ReplyGeneration{Content: content, Model: c.model}, nil
}

func (c *Client) messages(req store.ReplyRequest) []chatMessage {
	messages := []chatMessage{{
		Role: "system",
		Content: "你是可靠客服 Agent。只能基于提供的知识证据回答；不得编造商品、政策、承诺或处理结果。" +
			"如果证据不足，应说明需要补充知识；如果存在投诉、法律风险或人工诉求，应建议转人工。",
	}}
	for _, item := range tailMessages(req.History, 8) {
		if item.Role != "user" && item.Role != "assistant" {
			continue
		}
		messages = append(messages, chatMessage{Role: item.Role, Content: item.Content})
	}
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: fmt.Sprintf("用户问题：%s\n\n知识证据：\n%s", req.Question, evidenceText(req.Evidence)),
	})
	return messages
}

func evidenceText(items []store.KnowledgeArticle) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("- [%s] %s：%s", item.ID, item.Title, item.Content))
	}
	return strings.Join(parts, "\n")
}

func tailMessages(items []store.Message, limit int) []store.Message {
	if len(items) <= limit {
		return items
	}
	return items[len(items)-limit:]
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
