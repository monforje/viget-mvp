package profile

import (
	"errors"
	"viget-mvp/internal/models"
)

func (s *InMemoryStorage) CreateUserProfile(profile *models.UserProfile) error {
	if profile == nil || profile.ID == "" {
		return errors.New("invalid profile")
	}
	return s.SaveUserProfile(profile)
}

func (s *InMemoryStorage) GetUserProfileByID(userID string) (*models.UserProfile, error) {
	profile := s.GetUserProfile(userID)
	if profile == nil {
		return nil, errors.New("user not found")
	}
	return profile, nil
}

func (s *InMemoryStorage) UpdateUserProfile(profile *models.UserProfile) error {
	if profile == nil || profile.ID == "" {
		return errors.New("invalid profile")
	}
	return s.SaveUserProfile(profile)
}

func (s *InMemoryStorage) DeleteUserProfile(userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.users[userID]; !ok {
		return errors.New("user not found")
	}
	delete(s.users, userID)
	return nil
}
