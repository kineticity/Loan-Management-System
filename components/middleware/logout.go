package middleware

import (
	"sync"
)

var (
	blacklistedTokens = make(map[string]struct{})
	mu               sync.Mutex
)

func BlacklistToken(token string) {
	mu.Lock()
	defer mu.Unlock()
	blacklistedTokens[token] = struct{}{}
}

func IsTokenBlacklisted(token string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, exists := blacklistedTokens[token]
	return exists
}
