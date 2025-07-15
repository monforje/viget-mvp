package profile

import (
	"errors"
	"viget-mvp/internal/models"
)

func (s *InMemoryStorage) CreateTask(task *models.TaskProfile) error {
	if task == nil || task.ID == "" {
		return errors.New("invalid task")
	}
	return s.SaveTask(task)
}

func (s *InMemoryStorage) GetTaskByID(taskID string) (*models.TaskProfile, error) {
	task := s.GetTask(taskID)
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (s *InMemoryStorage) UpdateTask(task *models.TaskProfile) error {
	if task == nil || task.ID == "" {
		return errors.New("invalid task")
	}
	return s.SaveTask(task)
}

func (s *InMemoryStorage) DeleteTask(taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.tasks[taskID]; !ok {
		return errors.New("task not found")
	}
	delete(s.tasks, taskID)
	return nil
}

func (s *InMemoryStorage) ListTasks() []*models.TaskProfile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]*models.TaskProfile, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}
