package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) TokenManager {
	return TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m TokenManager) Generate(adminID int64) (string, time.Time, error) {
	expiresAt := m.now().Add(m.ttl)
	payload := fmt.Sprintf("%d:%d", adminID, expiresAt.Unix())
	signature := m.sign(payload)
	token := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature
	return token, expiresAt, nil
}

func (m TokenManager) Verify(token string) (int64, error) {
	payload, err := m.verifyPayload(token)
	if err != nil {
		return 0, err
	}

	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		return 0, ErrInvalidToken
	}

	adminID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || adminID <= 0 {
		return 0, ErrInvalidToken
	}

	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}

	if !m.now().Before(time.Unix(expiresUnix, 0)) {
		return 0, ErrInvalidToken
	}

	return adminID, nil
}

func (m TokenManager) verifyPayload(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrInvalidToken
	}

	payload := string(payloadBytes)
	expectedSignature := m.sign(payload)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return "", ErrInvalidToken
	}

	return payload, nil
}

func (m TokenManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
