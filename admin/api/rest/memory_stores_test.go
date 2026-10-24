package rest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/forgego/forge/admin/core"
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

// IssueToken implements TokenStore.
func (s *adminSessionStore) IssueToken(_ context.Context, username string, ttl time.Duration) (string, error) {
	return s.Issue(username, ttl)
}

// ValidateToken implements TokenStore.
func (s *adminSessionStore) ValidateToken(_ context.Context, token string) (string, bool, error) {
	session, ok := s.Validate(token)
	return session.Username, ok, nil
}

// RevokeToken implements TokenStore.
func (s *adminSessionStore) RevokeToken(_ context.Context, token string) error {
	s.Revoke(token)
	return nil
}

type savedViewStore struct {
	mu    sync.RWMutex
	views map[string]map[string][]SavedView
}

func newSavedViewStore() *savedViewStore {
	return &savedViewStore{views: make(map[string]map[string][]SavedView)}
}

func (s *savedViewStore) list(userID, model string) []SavedView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userViews, ok := s.views[userID]
	if !ok {
		return []SavedView{}
	}
	modelViews := userViews[model]
	if modelViews == nil {
		return []SavedView{}
	}
	return append([]SavedView{}, modelViews...)
}

func (s *savedViewStore) upsert(userID, model string, request SavedViewRequest) (SavedView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views[userID] == nil {
		s.views[userID] = make(map[string][]SavedView)
	}

	views := s.views[userID][model]
	for i, view := range views {
		if strings.EqualFold(view.Name, request.Name) {
			updated := view
			updated.Filters = request.Filters
			updated.Ordering = request.Ordering
			updated.Display = request.Display
			updated.UpdatedAt = time.Now()
			views[i] = updated
			s.views[userID][model] = views
			return updated, false
		}
	}

	newView := SavedView{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:      request.Name,
		Filters:   request.Filters,
		Ordering:  request.Ordering,
		Display:   request.Display,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.views[userID][model] = append(views, newView)
	return newView, true
}

func (s *savedViewStore) delete(userID, model, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	userViews, ok := s.views[userID]
	if !ok {
		return false
	}
	views := userViews[model]
	for i, view := range views {
		if view.ID == id {
			userViews[model] = append(views[:i], views[i+1:]...)
			return true
		}
	}
	return false
}

func (s *savedViewStore) ListSavedViews(_ context.Context, userKey, model string) ([]SavedView, error) {
	return s.list(userKey, model), nil
}

func (s *savedViewStore) SaveView(_ context.Context, userKey, model string, request SavedViewRequest) (SavedView, bool, error) {
	view, created := s.upsert(userKey, model, request)
	return view, created, nil
}

func (s *savedViewStore) DeleteSavedView(_ context.Context, userKey, model, id string) (bool, error) {
	return s.delete(userKey, model, id), nil
}

type loginLimiter struct {
	mu          sync.Mutex
	failures    map[string]*failureWindow
	max         int
	window      time.Duration
	now         func() time.Time
	lastCleanup time.Time
}

type failureWindow struct {
	count   int
	resetAt time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{failures: make(map[string]*failureWindow), max: 5,
		window: 15 * time.Minute, now: time.Now}
}

func (l *loginLimiter) getNow() time.Time {
	if l.now != nil {
		return l.now()
	}
	return time.Now()
}

func (l *loginLimiter) blocked(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.getNow()
	l.cleanupExpiredLocked(now)
	window, ok := l.failures[key]
	if !ok {
		return false, 0
	}
	if !now.Before(window.resetAt) {
		delete(l.failures, key)
		return false, 0
	}
	if window.count >= l.max {
		return true, window.resetAt.Sub(now)
	}
	return false, 0
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.getNow()
	l.cleanupExpiredLocked(now)
	window, ok := l.failures[key]
	if !ok || !now.Before(window.resetAt) {
		l.failures[key] = &failureWindow{count: 1, resetAt: now.Add(l.window)}
		return
	}
	window.count++
}

func (l *loginLimiter) cleanupExpiredLocked(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < time.Minute {
		return
	}
	for key, window := range l.failures {
		if !now.Before(window.resetAt) {
			delete(l.failures, key)
		}
	}
	l.lastCleanup = now
}

func (l *loginLimiter) success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

func (l *loginLimiter) Blocked(_ context.Context, key string) (bool, time.Duration, error) {
	blocked, remaining := l.blocked(key)
	return blocked, remaining, nil
}

func (l *loginLimiter) Failed(_ context.Context, key string) error {
	l.fail(key)
	return nil
}

func (l *loginLimiter) Succeeded(_ context.Context, key string) error {
	l.success(key)
	return nil
}

func testStores() Stores {
	return Stores{
		Tokens:        newAdminSessionStore(),
		SavedViews:    newSavedViewStore(),
		LoginAttempts: newLoginLimiter(),
	}
}

func newTestRouter(registry *core.Registry) *Router {
	return NewRouter(registry).SetStores(testStores())
}
