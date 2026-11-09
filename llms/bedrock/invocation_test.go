package bedrock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

type recordedRequest struct {
	path string
	body map[string]any
}

func newRecordingBedrockClient(t *testing.T) (*bedrockruntime.Client, *recordedRequest) {
	t.Helper()
	recorded := &recordedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded.path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&recorded.body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"message","role":"assistant","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	t.Cleanup(server.Close)

	client := bedrockruntime.New(bedrockruntime.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(server.URL),
		Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}, nil
		}),
	})
	return client, recorded
}

func TestGenerateContentInvocationTarget(t *testing.T) {
	const (
		profile  = "arn:aws:bedrock:us-east-1:123456789012:application-inference-profile/abc123"
		sonnet55 = "anthropic.claude-sonnet-5-5"
		haiku45  = "anthropic.claude-haiku-4-5"
	)
	tests := []struct {
		name         string
		opts         []Option
		callOptions  []llms.CallOption
		wantTarget   string
		wantThinking string
	}{
		{
			name:         "inference profile receives the request built for the model",
			opts:         []Option{WithModel(sonnet55), WithInferenceProfile(profile)},
			wantTarget:   profile,
			wantThinking: "between_tools",
		},
		{
			name:         "model receives the request without an inference profile",
			opts:         []Option{WithModel(sonnet55)},
			wantTarget:   sonnet55,
			wantThinking: "between_tools",
		},
		{
			name:         "per-call model replaces the model and skips the inference profile",
			opts:         []Option{WithModel(sonnet55), WithInferenceProfile(profile)},
			callOptions:  []llms.CallOption{llms.WithModel(haiku45)},
			wantTarget:   haiku45,
			wantThinking: "disabled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, recorded := newRecordingBedrockClient(t)
			llm, err := New(append(tt.opts, WithClient(client))...)
			require.NoError(t, err)

			callOptions := append([]llms.CallOption{llms.WithThinkingMode(llms.ThinkingModeNone)}, tt.callOptions...)
			_, err = llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "Hi")}, callOptions...)
			require.NoError(t, err)

			assert.Equal(t, "/model/"+tt.wantTarget+"/invoke", recorded.path)
			assert.Equal(t, map[string]any{"type": tt.wantThinking}, recorded.body["thinking"])
		})
	}
}
