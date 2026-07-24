package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// FIXED (was CWE-798, Hardcoded Credentials)
//
// Before: these were `const` string literals committed to source control --
// anyone with read access to the repo (or its git history) could read the
// signing key and forge tokens, or read the API key and use it directly.
//
// After: both are loaded from the environment at startup via MustLoadSecrets,
// which refuses to let the app start at all if either is missing, so there's
// no silent fallback to an empty or default key. In production these would
// come from a secrets manager (AWS Secrets Manager, Vault, etc.) rather than
// a bare env var -- this is the minimum viable improvement, not the ceiling.
var (
	jwtSecret     string
	stripeAPIKey  string
)

// MustLoadSecrets reads required secrets from the environment. Call this
// once at startup, before anything else. It calls log.Fatal (not a returned
// error) on purpose: an app with a missing signing key must never come up
// and start accepting requests.
func MustLoadSecrets() {
	jwtSecret = os.Getenv("JWT_SIGNING_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SIGNING_SECRET is not set; refusing to start")
	}

	stripeAPIKey = os.Getenv("STRIPE_API_KEY")
	if stripeAPIKey == "" {
		log.Fatal("STRIPE_API_KEY is not set; refusing to start")
	}
}

type Claims struct {
	Sub   string `json:"sub"`
	Admin bool   `json:"admin"`
	Exp   int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func b64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func b64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// IssueToken signs a normal HS256 token. The vulnerability isn't in issuance
// -- it's that VerifyToken (below) doesn't enforce the algorithm it expects.
func IssueToken(username string, isAdmin bool) (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, _ := json.Marshal(header)

	claims := Claims{Sub: username, Admin: isAdmin, Exp: time.Now().Add(1 * time.Hour).Unix()}
	claimsJSON, _ := json.Marshal(claims)

	signingInput := b64url(headerJSON) + "." + b64url(claimsJSON)

	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)

	return signingInput + "." + b64url(sig), nil
}

// allowedAlgs is the server's own allowlist of acceptable signing
// algorithms. VerifyToken checks the token's declared alg against this --
// the token's opinion is never authoritative on its own.
var allowedAlgs = map[string]bool{
	"HS256": true,
}

// VerifyToken -- FIXED (was: insecure JWT validation / "alg confusion")
//
// Before: the function switched on header.Alg -- whatever the token itself
// claimed -- and had a "none" case that skipped signature verification
// entirely. An attacker could hand-craft a token with header {"alg":"none"}
// and any payload they liked (e.g. "admin": true) and it would be accepted
// without ever checking a signature.
//
// After: the token's declared alg is checked against a fixed, server-owned
// allowlist (allowedAlgs) before anything else happens. There is no "none"
// branch and no default pass-through -- an alg the server doesn't
// explicitly recognize is rejected outright, regardless of what the token
// claims about itself. The server decides how tokens are checked; the token
// doesn't get a vote.
func VerifyToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}

	headerBytes, err := b64urlDecode(parts[0])
	if err != nil {
		return nil, err
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}

	if !allowedAlgs[header.Alg] {
		return nil, fmt.Errorf("rejected: alg %q is not on the server's allowlist", header.Alg)
	}

	claimsBytes, err := b64urlDecode(parts[1])
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	// Only HS256 is on the allowlist today, so this is the only branch
	// reachable past the check above -- but it's written as an explicit
	// switch (rather than "else assume HS256") so that adding a second
	// algorithm later requires deliberately adding its verification logic
	// here too, not just adding it to the map.
	switch header.Alg {
	case "HS256":
		signingInput := parts[0] + "." + parts[1]
		mac := hmac.New(sha256.New, []byte(jwtSecret))
		mac.Write([]byte(signingInput))
		expected := mac.Sum(nil)
		actual, err := b64urlDecode(parts[2])
		if err != nil {
			return nil, err
		}
		if !hmac.Equal(expected, actual) {
			return nil, errors.New("invalid signature")
		}
		return &claims, nil
	default:
		// Unreachable given allowedAlgs above, but kept as a hard failure
		// rather than falling through, in case the allowlist and this
		// switch ever drift out of sync.
		return nil, fmt.Errorf("no verification logic for alg %q", header.Alg)
	}
}
