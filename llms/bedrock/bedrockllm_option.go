package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/callbacks"
)

// Option is an option for the Bedrock LLM.
type Option func(*options)

type options struct {
	modelProvider    string
	modelID          string
	inferenceProfile string
	client           *bedrockruntime.Client
	callbackHandler  callbacks.Handler
}

// WithModel sets the ID of the model, such as "anthropic.claude-sonnet-5-5".
//
// The model ID decides how requests are built and, unless WithModelProvider is set, which provider handles them. Requests are sent to this model unless WithInferenceProfile is set. If not set, ModelAmazonTitanTextLiteV1 is used.
func WithModel(modelID string) Option {
	return func(o *options) {
		o.modelID = modelID
	}
}

// WithInferenceProfile sets the ID or ARN of the inference profile that requests are sent to.
//
// When set, requests go to the inference profile instead of the model set with WithModel. The model set with WithModel and the provider set with WithModelProvider still decide how requests are built, so the caller must keep them consistent with the model the inference profile routes to.
func WithInferenceProfile(inferenceProfile string) Option {
	return func(o *options) {
		o.inferenceProfile = inferenceProfile
	}
}

// WithModelProvider sets the provider of the model, such as "anthropic" or "meta".
//
// If not set, the provider is derived from the model ID. Set it when the model ID does not name the provider.
func WithModelProvider(modelProvider string) Option {
	return func(o *options) {
		o.modelProvider = modelProvider
	}
}

// WithClient allows setting a custom bedrockruntime.Client.
//
// You may use this to pass a custom bedrockruntime.Client
// with custom configuration options
// such as setting custom credentials, region, endpoint, etc.
//
// By default, a new client will be created using the default credentials chain.
func WithClient(client *bedrockruntime.Client) Option {
	return func(o *options) {
		o.client = client
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) Option {
	return func(o *options) {
		o.callbackHandler = callbackHandler
	}
}
