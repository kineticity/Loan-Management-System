package middleware

import (
	"sync"
)

var (
	blacklistedTokens = make(map[string]struct{})
	mu               sync.Mutex
)

// BlacklistToken adds a token to the blacklist
func BlacklistToken(token string) {
	mu.Lock()
	defer mu.Unlock()
	blacklistedTokens[token] = struct{}{}
}

// IsTokenBlacklisted checks if a token is blacklisted
func IsTokenBlacklisted(token string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, exists := blacklistedTokens[token]
	return exists
}
