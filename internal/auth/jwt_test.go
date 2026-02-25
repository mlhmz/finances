package auth

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret"

// ─── JWT tests ────────────────────────────────────────────────────────────────

func TestIssueAndParseAccessToken(t *testing.T) {
	token, err := IssueAccessToken("user-123", "user@example.com", testSecret, 3600)
	if err != nil {
		t.Fatalf("IssueAccessToken error: %v", err)
	}

	claims, err := ParseAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q; want %q", claims.Subject, "user-123")
	}
	if claims.Email != "user@example.com" {
		t.Errorf("Email = %q; want %q", claims.Email, "user@example.com")
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	token, _ := IssueAccessToken("user-123", "user@example.com", testSecret, 3600)
	_, err := ParseAccessToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	// ttl = -1 → already expired
	token, err := IssueAccessToken("user-999", "x@x.com", testSecret, -1)
	if err != nil {
		t.Fatalf("IssueAccessToken error: %v", err)
	}
	// Give the library a moment to recognise expiry
	time.Sleep(10 * time.Millisecond)
	_, err = ParseAccessToken(token, testSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestIssueAndParseRefreshToken(t *testing.T) {
	token, err := IssueRefreshToken("user-456", testSecret, 604800)
	if err != nil {
		t.Fatalf("IssueRefreshToken error: %v", err)
	}

	claims, err := ParseRefreshToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseRefreshToken error: %v", err)
	}
	if claims.Subject != "user-456" {
		t.Errorf("Subject = %q; want %q", claims.Subject, "user-456")
	}
	if claims.Type != "refresh" {
		t.Errorf("Type = %q; want %q", claims.Type, "refresh")
	}
}

func TestParseRefreshToken_AccessTokenRejected(t *testing.T) {
	// An access token must not be accepted as a refresh token.
	accessToken, _ := IssueAccessToken("user-1", "u@u.com", testSecret, 3600)
	_, err := ParseRefreshToken(accessToken, testSecret)
	if err == nil {
		t.Error("expected error when using access token as refresh token, got nil")
	}
}

// ─── OTP tests ────────────────────────────────────────────────────────────────

func TestGenerateOTP_Format(t *testing.T) {
	for range 20 {
		plain, hash, err := GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP error: %v", err)
		}
		if len(plain) != 8 {
			t.Errorf("plaintext length = %d; want 8", len(plain))
		}
		if plain != strings.ToUpper(plain) {
			t.Errorf("plaintext %q is not uppercase", plain)
		}
		if len(hash) != 64 {
			t.Errorf("hash length = %d; want 64 (SHA-256 hex)", len(hash))
		}
		// hash of the same input must match
		if HashOTP(plain) != hash {
			t.Errorf("HashOTP(%q) != returned hash", plain)
		}
	}
}

func TestHashOTP_CaseInsensitive(t *testing.T) {
	if HashOTP("abcd1234") != HashOTP("ABCD1234") {
		t.Error("HashOTP should be case-insensitive")
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		plain, _, err := GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP error: %v", err)
		}
		if seen[plain] {
			t.Errorf("duplicate OTP generated: %q", plain)
		}
		seen[plain] = true
	}
}
