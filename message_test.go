package anthropic_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/liushuangls/go-anthropic"
	"github.com/liushuangls/go-anthropic/internal/test"
	"github.com/liushuangls/go-anthropic/internal/test/checks"
)

func TestMessages(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken(),
		anthropic.WithBaseURL(baseUrl),
	)
	resp, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			{Role: anthropic.RoleUser, Content: "What is your name?"},
		},
		MaxTokens: 1000,
	})
	if err != nil {
		t.Fatalf("CreateMessages error: %v", err)
	}

	t.Logf("CreateMessages resp: %+v", resp)
}

func TestMessagesTokenError(t *testing.T) {
	server := test.NewTestServer()
	server.RegisterHandler("/v1/messages", handleMessagesEndpoint)

	ts := server.AnthropicTestServer()
	ts.Start()
	defer ts.Close()

	baseUrl := ts.URL + "/v1"
	client := anthropic.NewClient(
		test.GetTestToken()+"1",
		anthropic.WithBaseURL(baseUrl),
	)
	_, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			{Role: anthropic.RoleUser, Content: "What is your name?"},
		},
		MaxTokens: 1000,
	})
	checks.HasError(t, err, "should error")

	var e *anthropic.RequestError
	if !errors.As(err, &e) {
		t.Log("should request error")
	}

	t.Logf("CreateMessages error: %s", err)
}

func handleMessagesEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var messagesReq anthropic.MessagesRequest
	if messagesReq, err = getMessagesRequest(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	res := anthropic.MessagesResponse{
		Type: "completion",
		ID:   strconv.Itoa(int(time.Now().Unix())),
		Role: anthropic.RoleAssistant,
		Content: []anthropic.MessagesContent{
			{Type: "text", Text: "hello"},
		},
		StopReason: "end_turn",
		Model:      messagesReq.Model,
		Usage: anthropic.MessagesUsage{
			InputTokens:  10,
			OutputTokens: 10,
		},
	}
	resBytes, _ = json.Marshal(res)
	_, _ = w.Write(resBytes)
}

func getMessagesRequest(r *http.Request) (req anthropic.MessagesRequest, err error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(reqBody, &req)
	if err != nil {
		return
	}
	return
}
