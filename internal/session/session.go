package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Session represents a signed user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Signature []byte                 `json:"-"`
}

// Manager handles session creation and validation
type Manager struct {
	secret []byte
}

// NewManager creates a new session manager
func NewManager(secret string) *Manager {
	return &Manager{
		secret: []byte(secret),
	}
}

// Create creates a new session
func (m *Manager) Create(userID string, data map[string]interface{}) (*Session, error) {
	session := &Session{
		ID:        generateID(),
		UserID:    userID,
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	session.Sign(m.secret)
	return session, nil
}

// Validate validates and decodes a session token
func (m *Manager) Validate(token string) (*Session, error) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding: %w", err)
	}

	// Split payload and signature
	if len(data) < sha256.Size {
		return nil, fmt.Errorf("token too short")
	}

	signature := data[len(data)-sha256.Size:]
	payload := data[:len(data)-sha256.Size]

	// Verify signature
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(payload)
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode session
	var session Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return nil, fmt.Errorf("invalid session data: %w", err)
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// Encode encodes a session to a token string
func (m *Manager) Encode(session *Session) (string, error) {
	payload, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	data := append(payload, session.Signature...)
	return base64.URLEncoding.EncodeToString(data), nil
}

// Sign signs the session with HMAC
func (s *Session) Sign(secret []byte) {
	payload, _ := json.Marshal(s)
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	s.Signature = mac.Sum(nil)
}

func generateID() string {
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
