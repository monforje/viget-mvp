package models

import "time"

type UserProfile struct {
	ID         string                `json:"id"`
	TelegramID int64                 `json:"telegram_id"`
	Name       string                `json:"name"`
	Skills     map[string]SkillLevel `json:"skills"`
	Interests  []string              `json:"interests"`
	Experience []Experience          `json:"experience"`
	SoftSkills []string              `json:"soft_skills"`
	Goals      []string              `json:"goals"`
	Verified   map[string]bool       `json:"verified"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

type SkillLevel struct {
	Name     string `json:"name"`
	Level    int    `json:"level"` // 1-5
	Verified bool   `json:"verified"`
	Source   string `json:"source"` // interview, task, etc.
}

type Experience struct {
	Company     string   `json:"company"`
	Position    string   `json:"position"`
	Duration    string   `json:"duration"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

type TaskProfile struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	RequiredSkills map[string]int `json:"required_skills"` // skill -> min_level
	Budget         int            `json:"budget"`
	Deadline       time.Time      `json:"deadline"`
	CreatedBy      string         `json:"created_by"`
	Status         string         `json:"status"` // open, assigned, completed
	CreatedAt      time.Time      `json:"created_at"`
}

type InterviewSession struct {
	UserID      int64                  `json:"user_id"`
	Type        string                 `json:"type"` // "profile", "task"
	CurrentStep int                    `json:"current_step"`
	Answers     map[string]interface{} `json:"answers"`
	Context     map[string]interface{} `json:"context"`
	StartedAt   time.Time              `json:"started_at"`
}

type MatchResult struct {
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Score     float64   `json:"score"`
	Reasons   []string  `json:"reasons"`
	CreatedAt time.Time `json:"created_at"`
}
