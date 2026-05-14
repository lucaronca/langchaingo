package bedrockclient

import "time"

// ttlFromDuration maps an llms.CacheControl.Duration to the Bedrock ttl string.
// Bedrock currently accepts only "5m" and "1h" as values for the ttl field on
// cache_control / cachePoint. Returns an empty string when the duration is
// zero so the field can be omitted (the service defaults to a 5-minute TTL).
func ttlFromDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return ""
	case d <= 5*time.Minute:
		return "5m"
	default:
		return "1h"
	}
}
