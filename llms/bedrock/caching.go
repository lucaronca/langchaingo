package bedrock

import (
	"time"

	"github.com/tmc/langchaingo/llms"
)

// EphemeralCache returns an ephemeral cache control for Bedrock prompt
// caching. The optional ttl argument controls the wire format:
//
//   - EphemeralCache()                  → {"type":"ephemeral"} (no ttl;
//     Bedrock applies the default 5-minute behavior; compatible with every
//     Claude model that supports prompt caching).
//   - EphemeralCache(5*time.Minute)     → {"type":"ephemeral","ttl":"5m"}
//   - EphemeralCache(time.Hour)         → {"type":"ephemeral","ttl":"1h"}
//
// The explicit ttl forms are only accepted by Claude Opus 4.5, Haiku 4.5
// and Sonnet 4.5; older Claude models reject any value of ttl with "Extra
// inputs are not permitted". When in doubt, omit the argument.
//
// When you mix 1h and 5m breakpoints in the same request, Bedrock requires
// the 1h entries to appear before the 5m entries — same constraint as the
// direct Anthropic API.
//
// Only the first ttl argument is considered; extra arguments are ignored.
func EphemeralCache(ttl ...time.Duration) *llms.CacheControl {
	cc := &llms.CacheControl{Type: "ephemeral"}
	if len(ttl) > 0 {
		cc.Duration = ttl[0]
	}
	return cc
}
