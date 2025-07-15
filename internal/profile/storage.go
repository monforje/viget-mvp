// internal/profile/storage.go
package profile

import (
	"sync"
	"time"

	"viget-mvp/internal/models"
)

type InMemoryStorage struct {
	users map[string]*models.UserProfile
	tasks map[string]*models.TaskProfile
	mutex sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	storage := &InMemoryStorage{
		users: make(map[string]*models.UserProfile),
		tasks: make(map[string]*models.TaskProfile),
	}

	// Добавляем тестовые задачи
	storage.addTestTasks()

	return storage
}

func (s *InMemoryStorage) SaveUserProfile(profile *models.UserProfile) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	profile.UpdatedAt = time.Now()
	s.users[profile.ID] = profile
	return nil
}

func (s *InMemoryStorage) GetUserProfile(userID string) *models.UserProfile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.users[userID]
}

func (s *InMemoryStorage) SaveTask(task *models.TaskProfile) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tasks[task.ID] = task
	return nil
}

func (s *InMemoryStorage) GetTask(taskID string) *models.TaskProfile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.tasks[taskID]
}

func (s *InMemoryStorage) GetAvailableTasks() []*models.TaskProfile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var tasks []*models.TaskProfile
	for _, task := range s.tasks {
		if task.Status == "open" {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (s *InMemoryStorage) addTestTasks() {
	testTasks := []*models.TaskProfile{
		{
			ID:          "task_1",
			Title:       "Создать простой веб-сайт на React",
			Description: "Нужно сделать лендинг для стартапа с современным дизайном",
			RequiredSkills: map[string]int{
				"React":      2,
				"JavaScript": 3,
				"CSS":        2,
			},
			Budget:    30000,
			Deadline:  time.Now().AddDate(0, 0, 14),
			CreatedBy: "client_1",
			Status:    "open",
			CreatedAt: time.Now(),
		},
		{
			ID:          "task_2",
			Title:       "Парсер данных на Python",
			Description: "Написать скрипт для сбора данных с сайтов и сохранения в БД",
			RequiredSkills: map[string]int{
				"Python":        3,
				"BeautifulSoup": 2,
				"SQL":           2,
			},
			Budget:    25000,
			Deadline:  time.Now().AddDate(0, 0, 7),
			CreatedBy: "client_2",
			Status:    "open",
			CreatedAt: time.Now(),
		},
		{
			ID:          "task_3",
			Title:       "Мобильное приложение на Flutter",
			Description: "Простое приложение для заметок с синхронизацией",
			RequiredSkills: map[string]int{
				"Flutter":  3,
				"Dart":     3,
				"Firebase": 2,
			},
			Budget:    80000,
			Deadline:  time.Now().AddDate(0, 1, 0),
			CreatedBy: "client_3",
			Status:    "open",
			CreatedAt: time.Now(),
		},
	}

	for _, task := range testTasks {
		s.tasks[task.ID] = task
	}
}
