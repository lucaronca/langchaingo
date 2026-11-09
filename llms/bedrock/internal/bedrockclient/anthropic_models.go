package bedrockclient

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// Ref: https://platform.claude.com/docs/en/build-with-claude/thinking-troubleshooting
// Also: https://platform.claude.com/docs/en/build-with-claude/effort

const (
	anthropicThinkingAdaptive     = "adaptive"
	anthropicThinkingBetweenTools = "between_tools"
	anthropicThinkingDisabled     = "disabled"
	anthropicThinkingEnabled      = "enabled"
)

var (
	allEfforts = []llms.ThinkingEffort{
		llms.ThinkingEffortLow, llms.ThinkingEffortMedium, llms.ThinkingEffortHigh,
		llms.ThinkingEffortXHigh, llms.ThinkingEffortMax,
	}
	effortsWithoutXHigh = []llms.ThinkingEffort{
		llms.ThinkingEffortLow, llms.ThinkingEffortMedium, llms.ThinkingEffortHigh,
		llms.ThinkingEffortMax,
	}
	effortsUpToHigh = []llms.ThinkingEffort{
		llms.ThinkingEffortLow, llms.ThinkingEffortMedium, llms.ThinkingEffortHigh,
	}
)

// anthropicModel describes the thinking types, effort levels and sampling
// parameters a Claude model accepts.
type anthropicModel struct {
	name                         string
	thinkingTypes                []string
	efforts                      []llms.ThinkingEffort
	highestEffortWithThinkingOff llms.ThinkingEffort
	sampling                     bool
}

var anthropicModels = []anthropicModel{
	{name: "claude-fable-5-1", thinkingTypes: []string{anthropicThinkingAdaptive}, efforts: allEfforts},
	{name: "claude-fable-5", thinkingTypes: []string{anthropicThinkingAdaptive}, efforts: allEfforts},
	{name: "claude-mythos-5-1", thinkingTypes: []string{anthropicThinkingAdaptive}, efforts: allEfforts},
	{name: "claude-mythos-5", thinkingTypes: []string{anthropicThinkingAdaptive}, efforts: allEfforts},
	{name: "claude-opus-5-5", thinkingTypes: []string{anthropicThinkingAdaptive}, efforts: allEfforts},
	{
		name:                         "claude-opus-5",
		thinkingTypes:                []string{anthropicThinkingAdaptive, anthropicThinkingDisabled},
		efforts:                      allEfforts,
		highestEffortWithThinkingOff: llms.ThinkingEffortHigh,
	},
	{name: "claude-opus-4-8", thinkingTypes: []string{anthropicThinkingAdaptive, anthropicThinkingDisabled}, efforts: allEfforts},
	{name: "claude-opus-4-7", thinkingTypes: []string{anthropicThinkingAdaptive, anthropicThinkingDisabled}, efforts: allEfforts},
	{
		name:          "claude-opus-4-6",
		thinkingTypes: []string{anthropicThinkingAdaptive, anthropicThinkingDisabled, anthropicThinkingEnabled},
		efforts:       effortsWithoutXHigh,
		sampling:      true,
	},
	{
		name:          "claude-opus-4-5",
		thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled},
		efforts:       effortsUpToHigh,
		sampling:      true,
	},
	{name: "claude-opus-4-1", thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled}, sampling: true},
	{name: "claude-opus-4", thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled}, sampling: true},
	{
		name:                         "claude-sonnet-5-5",
		thinkingTypes:                []string{anthropicThinkingAdaptive, anthropicThinkingBetweenTools},
		efforts:                      allEfforts,
		highestEffortWithThinkingOff: llms.ThinkingEffortHigh,
	},
	{name: "claude-sonnet-5", thinkingTypes: []string{anthropicThinkingAdaptive, anthropicThinkingDisabled}, efforts: allEfforts},
	{
		name:          "claude-sonnet-4-6",
		thinkingTypes: []string{anthropicThinkingAdaptive, anthropicThinkingDisabled, anthropicThinkingEnabled},
		efforts:       effortsWithoutXHigh,
		sampling:      true,
	},
	{name: "claude-sonnet-4-5", thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled}, sampling: true},
	{name: "claude-sonnet-4", thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled}, sampling: true},
	{
		name:                         "claude-haiku-5-5",
		thinkingTypes:                []string{anthropicThinkingAdaptive, anthropicThinkingDisabled},
		efforts:                      allEfforts,
		highestEffortWithThinkingOff: llms.ThinkingEffortHigh,
	},
	{name: "claude-haiku-4-5", thinkingTypes: []string{anthropicThinkingDisabled, anthropicThinkingEnabled}, sampling: true},
}

// findAnthropicModel returns the entry of anthropicModels that modelID names,
// and false when modelID names none of them.
func findAnthropicModel(modelID string) (anthropicModel, bool) {
	id := strings.ToLower(modelID)
	for _, model := range anthropicModels {
		if namesModel(id, model.name) {
			return model, true
		}
	}
	return anthropicModel{}, false
}

// namesModel reports whether id contains name followed by nothing, by a
// version suffix such as "-v1:0", or by a snapshot date such as "-20250929".
func namesModel(id, name string) bool {
	i := strings.Index(id, name)
	if i < 0 {
		return false
	}
	rest := id[i+len(name):]
	switch {
	case rest == "", strings.HasPrefix(rest, ":"), strings.HasPrefix(rest, "-v"):
		return true
	case len(rest) >= 9 && rest[0] == '-' && strings.Trim(rest[1:9], "0123456789") == "":
		return true
	default:
		return false
	}
}

// thinkingOffType returns the thinking.type value that turns thinking off,
// and false when the model cannot turn thinking off.
func (m anthropicModel) thinkingOffType() (string, bool) {
	for _, thinkingType := range []string{anthropicThinkingBetweenTools, anthropicThinkingDisabled} {
		if slices.Contains(m.thinkingTypes, thinkingType) {
			return thinkingType, true
		}
	}
	return "", false
}

// applyAnthropicModelRules sets the thinking, effort and sampling fields of
// input to what the model named by modelID accepts, and leaves input unchanged
// for a model it does not know. It returns an error when the model cannot
// honor the requested thinking and effort.
func applyAnthropicModelRules(input *anthropicTextGenerationInput, modelID string, options llms.CallOptions) error {
	model, ok := findAnthropicModel(modelID)
	if !ok {
		return nil
	}
	if !model.sampling {
		input.Temperature, input.TopP, input.TopK = 0, 0, 0
	}

	config := llms.GetThinkingConfig(&options)
	if config == nil {
		return nil
	}
	thinkingOff := config.Mode == llms.ThinkingModeNone
	if thinkingOff {
		thinkingType, ok := model.thinkingOffType()
		if !ok {
			return llms.NewError(llms.ErrCodeInvalidRequest, "bedrock",
				fmt.Sprintf("%s cannot turn thinking off", model.name))
		}
		input.Thinking = &anthropicThinking{Type: thinkingType}
	}

	if config.Effort == "" || len(model.efforts) == 0 {
		return nil
	}
	if !slices.Contains(model.efforts, config.Effort) {
		return llms.NewError(llms.ErrCodeInvalidRequest, "bedrock",
			fmt.Sprintf("%s does not accept effort %q", model.name, config.Effort))
	}
	if thinkingOff && model.highestEffortWithThinkingOff != "" &&
		slices.Index(allEfforts, config.Effort) > slices.Index(allEfforts, model.highestEffortWithThinkingOff) {
		return llms.NewError(llms.ErrCodeInvalidRequest, "bedrock",
			fmt.Sprintf("%s does not accept effort %q while thinking is off", model.name, config.Effort))
	}
	input.OutputConfig = &anthropicOutputConfig{Effort: string(config.Effort)}
	return nil
}
