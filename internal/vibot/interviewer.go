// internal/vibot/interviewer.go
package vibot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"viget-mvp/internal/models"
	"viget-mvp/pkg/gpt"
)

type Interviewer struct {
	gptClient *gpt.Client
	sessions  map[int64]*models.InterviewSession
	mutex     sync.RWMutex
	questions *QuestionBank
}

func NewInterviewer(gptClient *gpt.Client) *Interviewer {
	return &Interviewer{
		gptClient: gptClient,
		sessions:  make(map[int64]*models.InterviewSession),
		questions: NewQuestionBank(),
	}
}

func (i *Interviewer) StartInterview(userID int64, interviewType string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Проверяем валидность типа интервью
	if interviewType != "profile" && interviewType != "task" {
		return fmt.Errorf("invalid interview type: %s", interviewType)
	}

	session := &models.InterviewSession{
		UserID:      userID,
		Type:        interviewType,
		CurrentStep: 0,
		Answers:     make(map[string]interface{}),
		Context:     make(map[string]interface{}),
		StartedAt:   time.Now(),
	}

	i.sessions[userID] = session
	return nil
}

func (i *Interviewer) GetCurrentQuestion(userID int64) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	session, exists := i.sessions[userID]
	if !exists {
		return "❌ Интервью не найдено. Используйте /interview для создания профиля или /create_task для создания задачи."
	}

	question := i.questions.GetQuestion(session.Type, session.CurrentStep, session.Context)

	// Добавляем префикс в зависимости от типа интервью
	var prefix string
	switch session.Type {
	case "profile":
		prefix = fmt.Sprintf("👤 Создание профиля (вопрос %d/%d)\n\n",
			session.CurrentStep+1, i.questions.GetMaxSteps(session.Type))
	case "task":
		prefix = fmt.Sprintf("📋 Создание задачи (вопрос %d/%d)\n\n",
			session.CurrentStep+1, i.questions.GetMaxSteps(session.Type))
	}

	return prefix + question
}

func (i *Interviewer) ProcessAnswer(userID int64, answer string) (string, bool, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	session, exists := i.sessions[userID]
	if !exists {
		return "", false, fmt.Errorf("session not found")
	}

	// Валидация ответа
	if strings.TrimSpace(answer) == "" {
		return "⚠️ Пожалуйста, дайте ответ на вопрос.", false, nil
	}

	// Сохраняем ответ
	questionKey := fmt.Sprintf("q_%d", session.CurrentStep)
	session.Answers[questionKey] = answer

	// Анализируем ответ с помощью GPT для контекста
	context, err := i.analyzeAnswer(answer, session)
	if err == nil {
		for k, v := range context {
			session.Context[k] = v
		}
	}

	// Переходим к следующему шагу
	session.CurrentStep++

	// Проверяем, завершено ли интервью
	maxSteps := i.questions.GetMaxSteps(session.Type)
	if session.CurrentStep >= maxSteps {
		return "", true, nil
	}

	// Возвращаем следующий вопрос
	nextQuestion := i.GetCurrentQuestion(userID)
	return nextQuestion, false, nil
}

func (i *Interviewer) ExtractProfile(userID int64) (*models.UserProfile, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	session, exists := i.sessions[userID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if session.Type != "profile" {
		return nil, fmt.Errorf("not a profile interview session")
	}

	// Собираем все ответы в один текст
	var allAnswers string
	for j := 0; j < session.CurrentStep; j++ {
		questionKey := fmt.Sprintf("q_%d", j)
		if answer, ok := session.Answers[questionKey]; ok {
			allAnswers += fmt.Sprintf("Вопрос %d: %v\n\n", j+1, answer)
		}
	}

	// Извлекаем структурированные данные через GPT
	extractedData, err := i.extractStructuredData(allAnswers, session.Type)
	if err != nil {
		return nil, err
	}

	// Создаем профиль
	profile := &models.UserProfile{
		ID:         strconv.FormatInt(userID, 10),
		TelegramID: userID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Skills:     make(map[string]models.SkillLevel),
		Verified:   make(map[string]bool),
	}

	// Заполняем данные из извлеченной информации
	if name, ok := extractedData["name"].(string); ok {
		profile.Name = name
	}

	if skills, ok := extractedData["skills"].(map[string]interface{}); ok {
		for skillName, levelData := range skills {
			if skillInfo, ok := levelData.(map[string]interface{}); ok {
				level := 1
				if l, ok := skillInfo["level"].(float64); ok {
					level = int(l)
				}
				profile.Skills[skillName] = models.SkillLevel{
					Name:     skillName,
					Level:    level,
					Verified: false,
					Source:   "interview",
				}
			}
		}
	}

	if interests, ok := extractedData["interests"].([]interface{}); ok {
		for _, interest := range interests {
			if str, ok := interest.(string); ok {
				profile.Interests = append(profile.Interests, str)
			}
		}
	}

	if goals, ok := extractedData["goals"].([]interface{}); ok {
		for _, goal := range goals {
			if str, ok := goal.(string); ok {
				profile.Goals = append(profile.Goals, str)
			}
		}
	}

	if softSkills, ok := extractedData["soft_skills"].([]interface{}); ok {
		for _, skill := range softSkills {
			if str, ok := skill.(string); ok {
				profile.SoftSkills = append(profile.SoftSkills, str)
			}
		}
	}

	// Удаляем сессию
	delete(i.sessions, userID)
	return profile, nil
}

func (i *Interviewer) ExtractTask(userID int64) (*models.TaskProfile, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	session, exists := i.sessions[userID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Собираем все ответы в один текст
	var allAnswers string
	for j := 0; j < session.CurrentStep; j++ {
		questionKey := fmt.Sprintf("q_%d", j)
		if answer, ok := session.Answers[questionKey]; ok {
			question := i.questions.GetQuestion(session.Type, j, session.Context)
			allAnswers += fmt.Sprintf("Q: %s\nA: %v\n\n", question, answer)
		}
	}

	extractedData, err := i.extractStructuredData(allAnswers, session.Type)
	if err != nil {
		return nil, err
	}

	if session.Type == "task" {
		task := &models.TaskProfile{
			ID:        fmt.Sprintf("task_%d", userID),
			CreatedBy: fmt.Sprintf("user_%d", userID),
			Status:    "open",
			CreatedAt: session.StartedAt,
		}

		if title, ok := extractedData["title"].(string); ok {
			task.Title = title
		}
		if desc, ok := extractedData["description"].(string); ok {
			task.Description = desc
		}
		if skills, ok := extractedData["required_skills"].(map[string]interface{}); ok {
			task.RequiredSkills = make(map[string]int)
			for skill, level := range skills {
				if lvl, ok := level.(float64); ok {
					task.RequiredSkills[skill] = int(lvl)
				}
			}
		}
		if budget, ok := extractedData["budget"].(float64); ok {
			task.Budget = int(budget)
		}
		if deadlineDays, ok := extractedData["deadline_days"].(float64); ok {
			task.Deadline = session.StartedAt.AddDate(0, 0, int(deadlineDays))
		}

		delete(i.sessions, userID)
		return task, nil
	}

	return nil, fmt.Errorf("unsupported interview type: %s", session.Type)
}

func (i *Interviewer) analyzeAnswer(answer string, session *models.InterviewSession) (map[string]interface{}, error) {
	var prompt string

	if session.Type == "profile" {
		prompt = fmt.Sprintf(`Проанализируй ответ пользователя на интервью для создания профиля и извлеки ключевую информацию.

Ответ: "%s"

Определи:
1. Основные навыки или технологии, упомянутые в ответе
2. Уровень опыта (junior/middle/senior)
3. Интересы и предпочтения
4. Любую другую важную информацию для профиля

Верни в JSON формате:
{
  "mentioned_skills": ["skill1", "skill2"],
  "experience_level": "junior|middle|senior",
  "interests": ["interest1"],
  "key_info": "краткое резюме"
}`, answer)
	} else if session.Type == "task" {
		prompt = fmt.Sprintf(`Проанализируй ответ пользователя на интервью для создания задачи и извлеки ключевую информацию.

Ответ: "%s"

Определи:
1. Упомянутые технологии или требования
2. Сложность задачи (simple/medium/complex)
3. Тип проекта
4. Любую другую важную информацию для задачи

Верни в JSON формате:
{
  "mentioned_technologies": ["tech1", "tech2"],
  "task_complexity": "simple|medium|complex",
  "project_type": "web|mobile|data|design|other",
  "key_info": "краткое резюме"
}`, answer)
	}

	response, err := i.gptClient.SendRequest(prompt)
	if err != nil {
		return nil, err
	}

	var context map[string]interface{}
	err = json.Unmarshal([]byte(response), &context)
	return context, err
}

func (i *Interviewer) extractStructuredData(answers string, sessionType string) (map[string]interface{}, error) {
	var prompt string

	if sessionType == "profile" {
		prompt = fmt.Sprintf(`Проанализируй интервью с пользователем для создания профиля и извлеки структурированную информацию.

Ответы на интервью:
%s

Извлеки и структурируй следующую информацию:
1. Имя пользователя
2. Технические навыки с уровнем (1-5)
3. Soft skills
4. Интересы и хобби
5. Профессиональные цели
6. Опыт работы

Верни в JSON формате:
{
  "name": "Имя",
  "skills": {
    "Python": {"level": 3, "confidence": 0.8},
    "JavaScript": {"level": 2, "confidence": 0.6}
  },
  "soft_skills": ["коммуникация", "командная работа"],
  "interests": ["машинное обучение", "веб-разработка"],
  "goals": ["стать senior разработчиком", "изучить Go"],
  "experience": [
    {
      "company": "ООО Пример",
      "position": "Junior Developer", 
      "duration": "6 месяцев",
      "skills": ["Python", "Django"]
    }
  ]
}`, answers)
	} else if sessionType == "task" {
		prompt = fmt.Sprintf(`Проанализируй интервью с пользователем для создания задачи и извлеки требования.

Ответы на интервью:
%s

Извлеки:
1. Название задачи
2. Подробное описание
3. Требуемые навыки с минимальным уровнем (1-5)
4. Бюджет
5. Сроки выполнения в днях

Верни в JSON формате:
{
  "title": "Название задачи",
  "description": "Подробное описание что нужно сделать",
  "required_skills": {
    "Python": 3,
    "React": 2,
    "CSS": 2
  },
  "budget": 50000,
  "deadline_days": 14
}`, answers)
	}

	response, err := i.gptClient.SendRequest(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(response), &result)
	return result, err
}

func (i *Interviewer) IsInInterview(userID int64) bool {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	_, exists := i.sessions[userID]
	return exists
}

func (i *Interviewer) GetInterviewType(userID int64) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	if session, exists := i.sessions[userID]; exists {
		return session.Type
	}
	return ""
}

func (i *Interviewer) CancelInterview(userID int64) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	delete(i.sessions, userID)
}
