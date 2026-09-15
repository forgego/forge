package rest

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

type adminSession struct {
	Username  string
	ExpiresAt time.Time
	Active    bool
}

type adminSessionStore struct {
	mu            sync.RWMutex
	sessions      map[string]adminSession
	lastPurge     time.Time
	purgeInterval time.Duration
}

func (s *adminSessionStore) purgeExpiredLocked(now time.Time) {
	for k, session := range s.sessions {
		if !session.Active || now.After(session.ExpiresAt) {
			delete(s.sessions, k)
		}
	}
}

func hashSessionToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func newAdminSessionStore() *adminSessionStore {
	return &adminSessionStore{
		sessions: make(map[string]adminSession),
	}
}

func (s *adminSessionStore) Issue(username string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", errors.New("invalid session ttl")
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(randomBytes)
	tokenHash := hashSessionToken(token)

	s.mu.Lock()
	now := time.Now()
	interval := s.purgeInterval
	if interval == 0 {
		interval = time.Minute
	}
	if s.lastPurge.IsZero() || now.Sub(s.lastPurge) >= interval {
		if len(s.sessions) > 0 {
			s.purgeExpiredLocked(now)
			s.lastPurge = now
		}
	}
	s.sessions[tokenHash] = adminSession{
		Username:  username,
		ExpiresAt: now.Add(ttl),
		Active:    true,
	}
	s.mu.Unlock()

	return token, nil
}

func (s *adminSessionStore) Validate(token string) (adminSession, bool) {
	tokenHash := hashSessionToken(token)

	s.mu.RLock()
	session, ok := s.sessions[tokenHash]
	s.mu.RUnlock()
	if !ok {
		return adminSession{}, false
	}

	if !session.Active {
		return adminSession{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, tokenHash)
		s.mu.Unlock()
		return adminSession{}, false
	}

	return session, true
}

func (s *adminSessionStore) Revoke(token string) bool {
	tokenHash := hashSessionToken(token)

	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[tokenHash]
	if !ok || !session.Active {
		return false
	}
	session.Active = false
	s.sessions[tokenHash] = session
	return true
}

func bearerToken(authorizationHeader string) (string, error) {
	header := strings.TrimSpace(authorizationHeader)
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization scheme")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("missing bearer token")
	}
	return token, nil
}
