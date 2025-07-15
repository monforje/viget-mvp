package matcher

import (
	"math"
	"strings"
	"viget-mvp/internal/models"
)

type Scorer struct{}

func NewScorer() *Scorer {
	return &Scorer{}
}

func (s *Scorer) Score(user *models.UserProfile, task *models.TaskProfile, availability float64) float64 {
	skillScore := s.skillMatch(user, task)
	interestScore := s.interestMatch(user, task)
	availabilityScore := math.Max(0, math.Min(1, availability))

	// Веса: навыки 60%, интересы 20%, доступность 20%
	total := skillScore*0.6 + interestScore*0.2 + availabilityScore*0.2
	return math.Min(1.0, total)
}

func (s *Scorer) skillMatch(user *models.UserProfile, task *models.TaskProfile) float64 {
	if len(task.RequiredSkills) == 0 {
		return 0.5
	}
	var total float64
	var matched int
	for skill, minLevel := range task.RequiredSkills {
		if us, ok := user.Skills[skill]; ok {
			if us.Level >= minLevel {
				total += 1.0 + float64(us.Level-minLevel)*0.1
			} else {
				total += float64(us.Level) / float64(minLevel) * 0.7
			}
			matched++
		}
	}
	if matched == 0 {
		return 0
	}
	avg := total / float64(len(task.RequiredSkills))
	coverage := float64(matched) / float64(len(task.RequiredSkills))
	return avg * coverage
}

func (s *Scorer) interestMatch(user *models.UserProfile, task *models.TaskProfile) float64 {
	if len(user.Interests) == 0 {
		return 0.5
	}
	taskText := strings.ToLower(task.Title + " " + task.Description)
	var count int
	for _, interest := range user.Interests {
		if strings.Contains(taskText, strings.ToLower(interest)) {
			count++
		}
	}
	return float64(count) / float64(len(user.Interests))
}
