package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const sessionIDSize = 32

var ErrInvalidSessionToken = errors.New("invalid session token")

type sessionTokens struct {
	secret []byte
}

func newSessionTokens(secret string) *sessionTokens {
	return &sessionTokens{secret: []byte(secret)}
}

type sessionCredentials struct {
	cookie string
	hash   []byte
}

func (s *sessionTokens) New() (sessionCredentials, error) {
	token := make([]byte, sessionIDSize)
	if _, err := rand.Read(token); err != nil {
		return sessionCredentials{}, fmt.Errorf("generate session token: %w", err)
	}

	return sessionCredentials{
		cookie: base64.RawURLEncoding.EncodeToString(token),
		hash:   hashSessionToken(s.secret, token),
	}, nil
}

func (s *sessionTokens) Hash(cookie string) ([]byte, error) {
	token, err := decodeSessionToken(cookie)
	if err != nil {
		return nil, fmt.Errorf("decode session token: %w", err)
	}

	return hashSessionToken(s.secret, token), nil
}

func decodeSessionToken(cookie string) ([]byte, error) {
	token, err := base64.RawURLEncoding.Strict().DecodeString(cookie)
	if err != nil || len(token) != sessionIDSize {
		return nil, ErrInvalidSessionToken
	}

	return token, nil
}

func hashSessionToken(secret, token []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(token)
	return mac.Sum(nil)
}
