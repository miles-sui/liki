package session

import (
	"sync"
	"liki/internal/llm"

	"testing"
	"time"

)

func TestNewSession(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	if sess.ID == "" {
		t.Error("session ID must not be empty")
	}
	if sess.IsClosed() {
		t.Error("new session should not be closed")
	}
	if len(sess.Messages) != 0 {
		t.Errorf("new session has %d messages, want 0", len(sess.Messages))
	}
	if time.Until(sess.ExpiresAt) < 4*time.Minute {
		t.Error("session expiry too soon")
	}
}

func TestGetSet(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	got, ok := s.Get(sess.ID)
	if !ok {
		t.Fatal("Get returned not found")
	}
	if got.ID != sess.ID {
		t.Errorf("Get: ID mismatch")
	}
}

func TestGetMissing(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("Get should return false for missing session")
	}
}

func TestDelete(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	s.Delete(sess.ID)
	_, ok := s.Get(sess.ID)
	if ok {
		t.Error("Get should return false after Delete")
	}
}

func TestCleanup(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	s.mu.Lock()
	s.sessions[sess.ID].ExpiresAt = time.Now().Add(-1 * time.Second)
	s.mu.Unlock()

	s.cleanup()

	s.mu.RLock()
	_, ok := s.sessions[sess.ID]
	s.mu.RUnlock()

	if ok {
		t.Error("expired session should have been cleaned up")
	}
}

func TestAppendMessage(t *testing.T) {
	sess := &Session{}
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "你好"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "你好！"})

	if len(sess.Messages) != 2 {
		t.Errorf("got %d messages, want 2", len(sess.Messages))
	}
	if sess.Messages[0].Role != llm.RoleUser {
		t.Errorf("first message role = %q, want user", sess.Messages[0].Role)
	}
}

func TestTouch(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	orig := sess.ExpiresAt

	s.Touch(sess.ID)

	got, _ := s.Get(sess.ID)
	if !got.ExpiresAt.After(orig) {
		t.Error("Touch should extend expiry")
	}
}

func TestTouchMissing(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	s.Touch("nonexistent")
}

func TestTruncateLocked_NoSystemMessage(t *testing.T) {
	sess := &Session{}
	for i := 0; i < 50; i++ {
		sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "msg"})
	}
	if len(sess.Messages) != maxMessages {
		t.Errorf("got %d messages, want %d (maxMessages)", len(sess.Messages), maxMessages)
	}
	if sess.Messages[0].Role == llm.RoleSystem {
		t.Error("first message should not be system")
	}
}

func TestTruncateLocked_WithSystemMessage(t *testing.T) {
	sess := &Session{}
	sess.AppendMessage(llm.Message{Role: llm.RoleSystem, Content: "system prompt"})
	for i := 0; i < 50; i++ {
		sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "msg"})
	}
	if len(sess.Messages) != maxMessages {
		t.Errorf("got %d messages, want %d (maxMessages)", len(sess.Messages), maxMessages)
	}
	if sess.Messages[0].Role != llm.RoleSystem {
		t.Error("first message should be system message preserved")
	}
}

func TestTruncateLocked_UnderLimit(t *testing.T) {
	sess := &Session{}
	sess.AppendMessage(llm.Message{Role: llm.RoleSystem, Content: "system"})
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "hello"})
	sess.AppendMessage(llm.Message{Role: llm.RoleAssistant, Content: "hi"})

	if len(sess.Messages) != 3 {
		t.Errorf("got %d messages, want 3 (no truncation needed)", len(sess.Messages))
	}
	if sess.Messages[0].Role != llm.RoleSystem {
		t.Error("system message should still be first")
	}
}

func TestSetPhase(t *testing.T) {
	sess := &Session{}
	sess.SetPhase(PhaseClosed)
	if !sess.IsClosed() {
		t.Error("session should be closed after SetPhase(PhaseClosed)")
	}
	sess.SetPhase(PhaseCollecting)
	if sess.IsClosed() {
		t.Error("session should not be closed after SetPhase(PhaseCollecting)")
	}
}

func TestSnapshotMessages(t *testing.T) {
	sess := &Session{}
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "hello"})
	msgs := sess.SnapshotMessages()
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	// Mutating the snapshot should not affect the session.
	msgs[0] = llm.Message{Role: llm.RoleAssistant, Content: "modified"}
	sess2 := sess.SnapshotMessages()
	if sess2[0].Role != llm.RoleUser {
		t.Error("snapshot modification should not affect session")
	}
}

func TestSetMessages(t *testing.T) {
	sess := &Session{}
	sess.AppendMessage(llm.Message{Role: llm.RoleSystem, Content: "sys"})
	for i := 0; i < 50; i++ {
		sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "msg"})
	}

	newMsgs := []llm.Message{
		{Role: llm.RoleUser, Content: "new"},
	}
	sess.SetMessages(newMsgs)
	if len(sess.Messages) != 1 {
		t.Errorf("got %d messages after SetMessages, want 1", len(sess.Messages))
	}
	if sess.Messages[0].Content != "new" {
		t.Errorf("content = %q, want 'new'", sess.Messages[0].Content)
	}
}

func TestSnapshot(t *testing.T) {
	sess := &Session{ID: "test-id", Phase: PhaseCollecting}
	sess.AppendMessage(llm.Message{Role: llm.RoleUser, Content: "hello"})

	id, phase, msgs := sess.Snapshot()
	if id != "test-id" {
		t.Errorf("id = %q, want test-id", id)
	}
	if phase != PhaseCollecting {
		t.Errorf("phase = %q, want collecting", phase)
	}
	if len(msgs) != 1 {
		t.Errorf("got %d messages, want 1", len(msgs))
	}
}

func TestDeleteExpired(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	sess := s.NewSession()
	s.mu.Lock()
	s.sessions[sess.ID].ExpiresAt = time.Now().Add(-1 * time.Second)
	s.mu.Unlock()

	// Get detects expired and calls deleteExpired.
	_, ok := s.Get(sess.ID)
	if ok {
		t.Error("Get should return false for expired session")
	}

	// Verify it's actually deleted.
	s.mu.RLock()
	_, exists := s.sessions[sess.ID]
	s.mu.RUnlock()
	if exists {
		t.Error("deleteExpired should have removed the session")
	}
}

func TestNewSession_AtCapacity(t *testing.T) {
	s := NewStore(5*time.Minute, 1)
	defer s.Stop()

	_ = s.NewSession() // fill to capacity
	sess := s.NewSession()
	if sess != nil {
		t.Error("NewSession should return nil when at capacity")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore(5 * time.Minute, 0)
	defer s.Stop()

	const numSessions = 20
	ids := make([]string, numSessions)
	for i := range numSessions {
		sess := s.NewSession()
		ids[i] = sess.ID
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := range goroutines {
		go func(goID int) {
			defer wg.Done()
			for i := range 200 {
				idx := (goID + i) % numSessions
				switch i % 5 {
				case 0:
					s.Get(ids[idx])
				case 1:
					s.Touch(ids[idx])
				case 2:
					s.NewSession()
				case 3:
					s.Get("nonexistent")
				case 4:
					s.Delete(ids[idx])
					s.NewSession()
				}
			}
		}(g)
	}

	wg.Wait()
}
