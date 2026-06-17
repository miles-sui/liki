// Package session provides in-memory chat session storage with TTL-based cleanup.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"

	"liki/internal/llm"
	"time"

)

// Phase represents a session lifecycle stage.
type Phase string

const (
	PhaseCollecting Phase = "collecting"
	PhaseClosed     Phase = "closed"
)

// Session holds the state of one agent chat conversation.
type Session struct {
	mu        sync.RWMutex
	ID        string
	Messages  []llm.Message
	Phase     Phase
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsClosed returns true when the session has completed.
func (s *Session) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Phase == PhaseClosed
}

const maxMessages = 40

// AppendMessage appends a message to the session history.
func (s *Session) AppendMessage(msg llm.Message) {
	s.mu.Lock()
	s.Messages = append(s.Messages, msg)
	s.truncateLocked()
	s.mu.Unlock()
}

// truncateLocked keeps the system message (if first) and the most recent messages
// up to maxMessages. Caller must hold s.mu.
func (s *Session) truncateLocked() {
	if len(s.Messages) <= maxMessages {
		return
	}
	keep := maxMessages
	offset := len(s.Messages) - keep
	if offset > 0 && s.Messages[0].Role == llm.RoleSystem {
		// Preserve system message + keep-1 recent messages.
		sys := s.Messages[0]
		s.Messages = append([]llm.Message{sys}, s.Messages[len(s.Messages)-keep+1:]...)
	} else {
		s.Messages = s.Messages[offset:]
	}
}

// SetPhase updates the session phase.
func (s *Session) SetPhase(phase Phase) {
	s.mu.Lock()
	s.Phase = phase
	s.mu.Unlock()
}

// SnapshotMessages returns a copy of the session's messages for safe concurrent reads.
func (s *Session) SnapshotMessages() []llm.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.copyMessagesLocked()
}

// SetMessages replaces the session's messages atomically.
func (s *Session) SetMessages(msgs []llm.Message) {
	s.mu.Lock()
	s.Messages = msgs
	s.truncateLocked()
	s.mu.Unlock()
}

// Snapshot returns a safe copy of the session's key fields.
func (s *Session) Snapshot() (id string, phase Phase, messages []llm.Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ID, s.Phase, s.copyMessagesLocked()
}

func (s *Session) copyMessagesLocked() []llm.Message {
	out := make([]llm.Message, len(s.Messages))
	copy(out, s.Messages)
	return out
}

// Store holds in-memory chat sessions with TTL cleanup.
type Store struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	ttl         time.Duration
	maxSessions int
	stopCh      chan struct{}
}

// NewStore creates a session store with the given TTL and maximum session count.
// maxSessions of 0 means unlimited. Call Stop() to shut down the cleanup goroutine.
func NewStore(ttl time.Duration, maxSessions int) *Store {
	s := &Store{
		sessions:    make(map[string]*Session),
		ttl:         ttl,
		maxSessions: maxSessions,
		stopCh:      make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

// NewSession creates a session with a random ID and returns it, or nil if the
// store is at capacity.
func (s *Store) NewSession() *Session {
	s.mu.Lock()
	if s.maxSessions > 0 && len(s.sessions) >= s.maxSessions {
		s.mu.Unlock()
		return nil
	}
	id := genID()
	sess := &Session{
		ID:        id,
		Phase:     PhaseCollecting,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.ttl),
	}
	s.sessions[id] = sess
	s.mu.Unlock()
	return sess
}

// Get returns a session by ID, or nil if not found or expired.
func (s *Store) Get(id string) (*Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	if !ok {
		s.mu.RUnlock()
		return nil, false
	}
	if time.Now().After(sess.ExpiresAt) {
		s.mu.RUnlock()
		s.deleteExpired(id)
		return nil, false
	}
	s.mu.RUnlock()
	return sess, true
}

// Touch extends the session's expiration time by the store's TTL.
func (s *Store) Touch(id string) {
	s.mu.Lock()
	if sess, ok := s.sessions[id]; ok {
		sess.ExpiresAt = time.Now().Add(s.ttl)
	}
	s.mu.Unlock()
}

// Delete removes a session from the store.
func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// deleteExpired removes the session only if it is currently expired.
func (s *Store) deleteExpired(id string) {
	s.mu.Lock()
	sess, ok := s.sessions[id]
	if ok && time.Now().After(sess.ExpiresAt) {
		delete(s.sessions, id)
	}
	s.mu.Unlock()
}

// Stop shuts down the cleanup goroutine.
func (s *Store) Stop() {
	close(s.stopCh)
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("session: panic in cleanup", "panic", r)
					}
				}()
				s.cleanup()
			}()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) cleanup() {
	now := time.Now()
	s.mu.Lock()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
}

func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		slog.Error("session: crypto/rand failed, falling back to time-based ID", "err", err)
		return hex.EncodeToString([]byte(fmt.Sprintf("%016x", time.Now().UnixNano())))
	}
	return hex.EncodeToString(b)
}
