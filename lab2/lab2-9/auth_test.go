package main

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

// TestMain ensures MustLoadSecrets has something to load, since VerifyToken
// and IssueToken both depend on jwtSecret being set.
func TestMain(m *testing.M) {
	os.Setenv("JWT_SIGNING_SECRET", "test-only-secret")
	os.Setenv("STRIPE_API_KEY", "sk_test_fake")
	MustLoadSecrets()
	os.Exit(m.Run())
}

func b64urlEnc(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// TestVerifyToken_RejectsAlgNone is the lab-required test: a forged token
// with header {"alg":"none"} and no real signature must be rejected, not
// silently accepted.
func TestVerifyToken_RejectsAlgNone(t *testing.T) {
	header := b64urlEnc(`{"alg":"none","typ":"JWT"}`)
	payload := b64urlEnc(`{"sub":"eve","admin":true,"exp":9999999999}`)
	forged := header + "." + payload + "." // empty signature segment

	claims, err := VerifyToken(forged)

	if err == nil {
		t.Fatalf("expected alg:none token to be rejected, but it was accepted with claims: %+v", claims)
	}
	if claims != nil {
		t.Fatalf("expected nil claims on rejection, got: %+v", claims)
	}
}

// TestVerifyToken_RejectsUnknownAlg makes sure the allowlist rejects
// anything it doesn't explicitly recognize, not just "none" specifically.
func TestVerifyToken_RejectsUnknownAlg(t *testing.T) {
	header := b64urlEnc(`{"alg":"RS256","typ":"JWT"}`)
	payload := b64urlEnc(`{"sub":"eve","admin":true,"exp":9999999999}`)
	forged := header + "." + payload + ".somefakesig"

	if _, err := VerifyToken(forged); err == nil {
		t.Fatal("expected alg:RS256 token to be rejected (not on allowlist), but it was accepted")
	}
}

// TestVerifyToken_AcceptsValidHS256 is the control case: a properly issued
// and signed token must still work after the fix.
func TestVerifyToken_AcceptsValidHS256(t *testing.T) {
	token, err := IssueToken("alice", true)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	claims, err := VerifyToken(token)
	if err != nil {
		t.Fatalf("expected valid HS256 token to be accepted, got error: %v", err)
	}
	if claims.Sub != "alice" || !claims.Admin {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

// TestVerifyToken_RejectsTamperedSignature ensures a valid alg with a
// tampered payload/signature still fails -- the allowlist fix shouldn't
// have weakened the actual signature check.
func TestVerifyToken_RejectsTamperedSignature(t *testing.T) {
	token, err := IssueToken("bob", false)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected a 3-part token, got %d parts", len(parts))
	}

	// Flip bob's admin claim to true without re-signing.
	tamperedPayload := b64urlEnc(`{"sub":"bob","admin":true,"exp":9999999999}`)
	tampered := parts[0] + "." + tamperedPayload + "." + parts[2]

	if _, err := VerifyToken(tampered); err == nil {
		t.Fatal("expected tampered token to be rejected due to signature mismatch")
	}
}
