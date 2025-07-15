package matcher

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"viget-mvp/internal/models"
)

type Matcher struct{}

func NewMatcher() *Matcher {
	return &Matcher{}
}

func (m *Matcher) FindMatchingTasks(user *models.UserProfile, tasks []*models.TaskProfile) []models.MatchResult {
	var matches []models.MatchResult

	for _, task := range tasks {
		if task.Status != "open" {
			continue
		}

		score := m.calculateMatchScore(user, task)
		reasons := m.generateMatchReasons(user, task, score)

		if score > 0.3 { // Минимальный порог совпадения
			matches = append(matches, models.MatchResult{
				TaskID:  task.ID,
				UserID:  user.ID,
				Score:   score,
				Reasons: reasons,
			})
		}
	}

	// Сортируем по убыванию score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}

func (m *Matcher) calculateMatchScore(user *models.UserProfile, task *models.TaskProfile) float64 {
	skillScore := m.calculateSkillMatch(user, task)
	interestScore := m.calculateInterestMatch(user, task)

	// Веса: навыки 80%, интересы 20%
	totalScore := skillScore*0.8 + interestScore*0.2

	return math.Min(1.0, totalScore)
}

func (m *Matcher) calculateSkillMatch(user *models.UserProfile, task *models.TaskProfile) float64 {
	if len(task.RequiredSkills) == 0 {
		return 0.5 // Нейтральная оценка, если требования не указаны
	}

	var totalScore float64
	var matchedSkills int

	for requiredSkill, minLevel := range task.RequiredSkills {
		userSkill, hasSkill := user.Skills[requiredSkill]

		if hasSkill {
			if userSkill.Level >= minLevel {
				// Бонус за превышение минимального уровня
				bonus := float64(userSkill.Level-minLevel) * 0.1
				skillScore := 1.0 + bonus
				totalScore += skillScore
			} else {
				// Частичное совпадение, если уровень ниже требуемого
				ratio := float64(userSkill.Level) / float64(minLevel)
				totalScore += ratio * 0.7
			}
			matchedSkills++
		}
		// Если навыка нет, добавляем 0 к общему счету
	}

	if matchedSkills == 0 {
		return 0
	}

	averageScore := totalScore / float64(len(task.RequiredSkills))

	// Штраф за отсутствующие навыки
	skillCoverage := float64(matchedSkills) / float64(len(task.RequiredSkills))

	return averageScore * skillCoverage
}

func (m *Matcher) calculateInterestMatch(user *models.UserProfile, task *models.TaskProfile) float64 {
	if len(user.Interests) == 0 {
		return 0.5
	}

	taskText := strings.ToLower(task.Title + " " + task.Description)
	var matchCount int

	for _, interest := range user.Interests {
		if strings.Contains(taskText, strings.ToLower(interest)) {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(user.Interests))
}

func (m *Matcher) generateMatchReasons(user *models.UserProfile, task *models.TaskProfile, score float64) []string {
	var reasons []string

	// Анализируем совпадения навыков
	for requiredSkill, minLevel := range task.RequiredSkills {
		if userSkill, hasSkill := user.Skills[requiredSkill]; hasSkill {
			if userSkill.Level >= minLevel {
				reasons = append(reasons, fmt.Sprintf("✅ %s: ваш уровень %d/%d", requiredSkill, userSkill.Level, minLevel))
			} else {
				reasons = append(reasons, fmt.Sprintf("⚠️ %s: ваш уровень %d/%d (ниже требуемого)", requiredSkill, userSkill.Level, minLevel))
			}
		} else {
			reasons = append(reasons, fmt.Sprintf("❌ %s: навык отсутствует", requiredSkill))
		}
	}

	// Анализируем совпадения интересов
	taskText := strings.ToLower(task.Title + " " + task.Description)
	for _, interest := range user.Interests {
		if strings.Contains(taskText, strings.ToLower(interest)) {
			reasons = append(reasons, fmt.Sprintf("💡 Совпадает с вашим интересом: %s", interest))
		}
	}

	// Общая оценка
	if score > 0.8 {
		reasons = append([]string{"🎯 Отличное совпадение!"}, reasons...)
	} else if score > 0.6 {
		reasons = append([]string{"👍 Хорошее совпадение"}, reasons...)
	} else if score > 0.4 {
		reasons = append([]string{"🤔 Частичное совпадение"}, reasons...)
	}

	return reasons
}

func (m *Matcher) RecommendTopTasks(user *models.UserProfile, tasks []*models.TaskProfile, topN int) []models.MatchResult {
	matches := m.FindMatchingTasks(user, tasks)
	if len(matches) > topN {
		return matches[:topN]
	}
	return matches
}
