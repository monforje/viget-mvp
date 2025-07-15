package vibot

import (
	"sync"
	"viget-mvp/internal/models"
)

type SessionStorage struct {
	mu       sync.RWMutex
	sessions map[int64]*models.InterviewSession
}

func NewSessionStorage() *SessionStorage {
	return &SessionStorage{
		sessions: make(map[int64]*models.InterviewSession),
	}
}

func (s *SessionStorage) Create(session *models.InterviewSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.UserID] = session
}

func (s *SessionStorage) Get(userID int64) (*models.InterviewSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[userID]
	return sess, ok
}

func (s *SessionStorage) Delete(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, userID)
}

func (s *SessionStorage) Exists(userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.sessions[userID]
	return ok
}
