// Package entities provides types and constants for the idempotency module.
package entities

import "time"

// IdempotencyRecord represents a cached response for an idempotency key.
type IdempotencyRecord struct {
	Key            string
	ResponseStatus int
	ResponseBody   []byte
	CreatedAt      time.Time
}
