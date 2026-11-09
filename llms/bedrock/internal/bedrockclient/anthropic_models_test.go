package bedrockclient

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestFindAnthropicModel(t *testing.T) {
	tests := []struct {
		modelID string
		want    string
	}{
		{modelID: "anthropic.claude-sonnet-5-5", want: "claude-sonnet-5-5"},
		{modelID: "us.anthropic.claude-sonnet-5-5", want: "claude-sonnet-5-5"},
		{modelID: "anthropic.claude-sonnet-5", want: "claude-sonnet-5"},
		{modelID: "anthropic.claude-sonnet-4-6", want: "claude-sonnet-4-6"},
		{modelID: "global.anthropic.claude-sonnet-4-5-20250929-v1:0", want: "claude-sonnet-4-5"},
		{modelID: "anthropic.claude-sonnet-4-20250514-v1:0", want: "claude-sonnet-4"},
		{modelID: "anthropic.claude-opus-4-6-v1", want: "claude-opus-4-6"},
		{modelID: "anthropic.claude-opus-4-1-20250805-v1:0", want: "claude-opus-4-1"},
		{modelID: "us.anthropic.claude-opus-5-5", want: "claude-opus-5-5"},
		{modelID: "anthropic.claude-haiku-4-5-20251001-v1:0", want: "claude-haiku-4-5"},
		{modelID: "arn:aws:bedrock:us-east-1:123456789012:application-inference-profile/abc123", want: ""},
		{modelID: "anthropic.claude-opus-5-1", want: ""},
		{modelID: "anthropic.claude-3-haiku-20240307-v1:0", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			model, ok := findAnthropicModel(tt.modelID)
			assert.Equal(t, tt.want != "", ok)
			assert.Equal(t, tt.want, model.name)
		})
	}
}

func thinkingOptions(mode llms.ThinkingMode, effort llms.ThinkingEffort) llms.CallOptions {
	var options llms.CallOptions
	llms.WithThinking(&llms.ThinkingConfig{Mode: mode, Effort: effort})(&options)
	return options
}

func requestBody(t *testing.T, input anthropicTextGenerationInput) map[string]any {
	t.Helper()
	data, err := json.Marshal(input)
	require.NoError(t, err)
	var body map[string]any
	require.NoError(t, json.Unmarshal(data, &body))
	return body
}

func TestApplyAnthropicModelRulesThinkingAndEffort(t *testing.T) {
	tests := []struct {
		name         string
		modelID      string
		options      llms.CallOptions
		wantThinking any
		wantOutput   any
	}{
		{
			name:         "sonnet 5.5 turns thinking off with between_tools",
			modelID:      "anthropic.claude-sonnet-5-5",
			options:      thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortMedium),
			wantThinking: map[string]any{"type": "between_tools"},
			wantOutput:   map[string]any{"effort": "medium"},
		},
		{
			name:         "sonnet 5 turns thinking off with disabled",
			modelID:      "anthropic.claude-sonnet-5",
			options:      thinkingOptions(llms.ThinkingModeNone, ""),
			wantThinking: map[string]any{"type": "disabled"},
		},
		{
			name:         "sonnet 4.5 drops the effort it does not take",
			modelID:      "anthropic.claude-sonnet-4-5-20250929-v1:0",
			options:      thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortMedium),
			wantThinking: map[string]any{"type": "disabled"},
		},
		{
			name:       "thinking left to the model sends only the effort",
			modelID:    "anthropic.claude-sonnet-5-5",
			options:    thinkingOptions(llms.ThinkingModeAuto, llms.ThinkingEffortXHigh),
			wantOutput: map[string]any{"effort": "xhigh"},
		},
		{
			name:         "haiku 5.5 accepts high effort with thinking off",
			modelID:      "anthropic.claude-haiku-5-5",
			options:      thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortHigh),
			wantThinking: map[string]any{"type": "disabled"},
			wantOutput:   map[string]any{"effort": "high"},
		},
		{
			name:    "no thinking config leaves both fields out",
			modelID: "anthropic.claude-sonnet-5-5",
		},
		{
			name:    "unknown model leaves both fields out",
			modelID: "arn:aws:bedrock:us-east-1:123456789012:application-inference-profile/abc123",
			options: thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortMedium),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input anthropicTextGenerationInput
			require.NoError(t, applyAnthropicModelRules(&input, tt.modelID, tt.options))
			body := requestBody(t, input)
			assert.Equal(t, tt.wantThinking, body["thinking"])
			assert.Equal(t, tt.wantOutput, body["output_config"])
		})
	}
}

func TestApplyAnthropicModelRulesRejectsWhatTheModelCannotHonor(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		options llms.CallOptions
		wantErr string
	}{
		{
			name:    "opus 5.5 cannot turn thinking off",
			modelID: "anthropic.claude-opus-5-5",
			options: thinkingOptions(llms.ThinkingModeNone, ""),
			wantErr: "claude-opus-5-5 cannot turn thinking off",
		},
		{
			name:    "sonnet 5.5 rejects xhigh with thinking off",
			modelID: "anthropic.claude-sonnet-5-5",
			options: thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortXHigh),
			wantErr: `claude-sonnet-5-5 does not accept effort "xhigh" while thinking is off`,
		},
		{
			name:    "haiku 5.5 rejects max with thinking off",
			modelID: "anthropic.claude-haiku-5-5",
			options: thinkingOptions(llms.ThinkingModeNone, llms.ThinkingEffortMax),
			wantErr: `claude-haiku-5-5 does not accept effort "max" while thinking is off`,
		},
		{
			name:    "sonnet 4.6 has no xhigh effort",
			modelID: "anthropic.claude-sonnet-4-6",
			options: thinkingOptions(llms.ThinkingModeAuto, llms.ThinkingEffortXHigh),
			wantErr: `claude-sonnet-4-6 does not accept effort "xhigh"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input anthropicTextGenerationInput
			err := applyAnthropicModelRules(&input, tt.modelID, tt.options)
			var llmErr *llms.Error
			require.ErrorAs(t, err, &llmErr)
			assert.Equal(t, llms.ErrCodeInvalidRequest, llmErr.Code)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestApplyAnthropicModelRulesSampling(t *testing.T) {
	tests := []struct {
		name         string
		modelID      string
		wantSampling bool
	}{
		{name: "sonnet 5.5 rejects sampling parameters", modelID: "anthropic.claude-sonnet-5-5"},
		{name: "sonnet 4.6 accepts sampling parameters", modelID: "anthropic.claude-sonnet-4-6", wantSampling: true},
		{name: "unknown model keeps sampling parameters", modelID: "anthropic.claude-3-haiku-20240307-v1:0", wantSampling: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := anthropicTextGenerationInput{Temperature: 0.7, TopP: 0.9, TopK: 5}
			require.NoError(t, applyAnthropicModelRules(&input, tt.modelID, llms.CallOptions{}))
			body := requestBody(t, input)
			for _, field := range []string{"temperature", "top_p", "top_k"} {
				_, present := body[field]
				assert.Equal(t, tt.wantSampling, present, field)
			}
		})
	}
}

func TestHandleAnthropicStreamingChunkForwardsOnlyTextDeltas(t *testing.T) {
	var chunks []string
	options := llms.CallOptions{StreamingFunc: func(_ context.Context, chunk []byte) error {
		chunks = append(chunks, string(chunk))
		return nil
	}}
	choice := &llms.ContentChoice{GenerationInfo: map[string]any{}}

	events := []string{
		`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Let me see."}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"abc"}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Hello"}}`,
		`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":" world"}}`,
	}
	for _, event := range events {
		var resp streamingCompletionResponseChunk
		require.NoError(t, json.Unmarshal([]byte(event), &resp))
		require.NoError(t, handleAnthropicStreamingChunk(context.Background(), resp, choice, options))
	}

	assert.Equal(t, []string{"Hello", " world"}, chunks)
	assert.Equal(t, "Hello world", choice.Content)
}
