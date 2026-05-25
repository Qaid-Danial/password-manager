package utils

import "golang.org/x/crypto/bcrypt"

// dummyHash is pre-computed at startup for use in constant-time login checks.
// When a login email is not found, we still run a bcrypt comparison against this
// hash so an attacker cannot enumerate valid emails by measuring response time.
var dummyHash []byte

func init() {
	// bcrypt.MinCost (4) is fast enough to not slow startup, while still
	// exercising the full bcrypt code path for timing attack prevention.
	h, _ := bcrypt.GenerateFromPassword([]byte("_dummy_"), bcrypt.MinCost)
	dummyHash = h
}

// HashPassword hashes a plaintext password with bcrypt at cost 12.
// Cost 12 takes ~250 ms on typical hardware — slow enough to make offline
// brute-force expensive while remaining acceptable for interactive login.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// bcrypt.CompareHashAndPassword is constant-time; no manual timing is needed.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// DummyPasswordCompare runs a bcrypt comparison against a pre-computed dummy
// hash. Call this when a login email is not found to maintain constant timing.
func DummyPasswordCompare(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}
