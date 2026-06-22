package auth

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
)

func TestSessionTokens_New(t *testing.T) {
	t.Parallel()

	tokens := newSessionTokens("test-secret")
	first, err := tokens.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decoded, err := base64.RawURLEncoding.Strict().DecodeString(first.cookie)
	if err != nil {
		t.Fatalf("decode generated cookie: %v", err)
	}
	if len(decoded) != sessionIDSize {
		t.Fatalf("decoded token length = %d, want %d", len(decoded), sessionIDSize)
	}
	if len(first.hash) != sha256.Size {
		t.Fatalf("hash length = %d, want %d", len(first.hash), sha256.Size)
	}

	hash, err := tokens.Hash(first.cookie)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !bytes.Equal(hash, first.hash) {
		t.Fatalf("Hash(cookie) does not match the hash returned by New()")
	}

	second, err := tokens.New()
	if err != nil {
		t.Fatalf("second New() error = %v", err)
	}
	if first.cookie == second.cookie {
		t.Fatal("consecutive New() calls returned the same cookie")
	}
	if bytes.Equal(first.hash, second.hash) {
		t.Fatal("consecutive New() calls returned the same hash")
	}
}

func TestSessionTokens_Hash(t *testing.T) {
	t.Parallel()

	rawToken := bytes.Repeat([]byte{0x42}, sessionIDSize)
	cookie := base64.RawURLEncoding.EncodeToString(rawToken)

	first := newSessionTokens("first-secret")
	firstHash, err := first.Hash(cookie)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	repeatedHash, err := first.Hash(cookie)
	if err != nil {
		t.Fatalf("repeated Hash() error = %v", err)
	}
	if !bytes.Equal(firstHash, repeatedHash) {
		t.Fatal("Hash() is not deterministic for the same secret and token")
	}

	second := newSessionTokens("second-secret")
	secondHash, err := second.Hash(cookie)
	if err != nil {
		t.Fatalf("Hash() with another secret error = %v", err)
	}
	if bytes.Equal(firstHash, secondHash) {
		t.Fatal("Hash() returned the same value for different secrets")
	}
}

func TestSessionTokens_HashRejectsInvalidCookies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		cookie string
	}{
		{
			name:   "empty",
			cookie: "",
		},
		{
			name:   "invalid base64url character",
			cookie: "!",
		},
		{
			name:   "too short",
			cookie: base64.RawURLEncoding.EncodeToString(make([]byte, sessionIDSize-1)),
		},
		{
			name:   "too long",
			cookie: base64.RawURLEncoding.EncodeToString(make([]byte, sessionIDSize+1)),
		},
		{
			name:   "padded encoding",
			cookie: base64.URLEncoding.EncodeToString(make([]byte, sessionIDSize)),
		},
	}

	tokens := newSessionTokens("test-secret")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tokens.Hash(tt.cookie)
			if !errors.Is(err, ErrInvalidSessionToken) {
				t.Fatalf("Hash(%q) error = %v, want %v", tt.cookie, err, ErrInvalidSessionToken)
			}
			if got, want := err.Error(), "decode session token: invalid session token"; got != want {
				t.Fatalf("Hash(%q) error = %q, want %q", tt.cookie, got, want)
			}
		})
	}
}
