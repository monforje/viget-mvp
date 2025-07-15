// internal/vibot/questions.go
package vibot

import (
	"fmt"
	"strings"
)

type QuestionBank struct {
	profileQuestions []QuestionTemplate
	taskQuestions    []QuestionTemplate
}

type QuestionTemplate struct {
	Text     string
	Required bool
	Type     string // "text", "number", "choice"
	Validate func(string) bool
}

func NewQuestionBank() *QuestionBank {
	return &QuestionBank{
		profileQuestions: []QuestionTemplate{
			{
				Text:     "👋 Привет! Давайте знакомиться. Как вас зовут?",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "💻 Расскажите о своем опыте в IT. Какими технологиями владеете? (например: Python, JavaScript, React)",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "📊 Оцените свой общий уровень в программировании от 1 до 5, где:\n1 - начинающий\n2 - базовые знания\n3 - уверенный пользователь\n4 - продвинутый\n5 - эксперт",
				Required: true,
				Type:     "number",
				Validate: func(s string) bool {
					return strings.Contains("12345", s) && len(s) == 1
				},
			},
			{
				Text:     "🎯 Что вас интересует в работе? Какие проекты хотели бы делать? (веб-разработка, мобильные приложения, данные, дизайн и т.д.)",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "🤝 Расскажите о своих сильных сторонах в работе. Что у вас получается особенно хорошо?",
				Required: false,
				Type:     "text",
			},
			{
				Text:     "🚀 Какие профессиональные цели хотите достичь в ближайшее время?",
				Required: false,
				Type:     "text",
			},
			{
				Text:     "💼 Есть ли у вас опыт удаленной работы или фриланса? Если да, расскажите кратко.",
				Required: false,
				Type:     "text",
			},
		},
		taskQuestions: []QuestionTemplate{
			{
				Text:     "📝 Как называется ваша задача? Придумайте краткое и понятное название.",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "📋 Опишите подробно, что нужно сделать. Какой результат вы ожидаете получить?",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "🛠️ Какие технические навыки нужны исполнителю? Укажите технологии и желаемый уровень (например: Python - 3/5, React - 2/5)",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "💰 Какой бюджет вы готовы выделить на эту задачу? Укажите сумму в рублях.",
				Required: true,
				Type:     "number",
			},
			{
				Text:     "⏰ В какие сроки нужно выполнить задачу? Укажите количество дней или конкретную дату.",
				Required: true,
				Type:     "text",
			},
			{
				Text:     "⭐ Есть ли особые требования к исполнителю? (опыт, портфолио, общение и т.д.)",
				Required: false,
				Type:     "text",
			},
		},
	}
}

func (q *QuestionBank) GetQuestion(interviewType string, step int, context map[string]interface{}) string {
	var questions []QuestionTemplate

	switch interviewType {
	case "profile":
		questions = q.profileQuestions
	case "task":
		questions = q.taskQuestions
	default:
		return "❌ Неизвестный тип интервью"
	}

	if step >= len(questions) {
		return "✅ Интервью завершено"
	}

	question := questions[step]

	// Адаптируем вопрос на основе контекста
	if len(context) > 0 {
		return q.adaptQuestion(question.Text, step, interviewType, context)
	}

	return question.Text
}

func (q *QuestionBank) GetMaxSteps(interviewType string) int {
	switch interviewType {
	case "profile":
		return len(q.profileQuestions)
	case "task":
		return len(q.taskQuestions)
	default:
		return 0
	}
}

func (q *QuestionBank) adaptQuestion(question string, step int, interviewType string, context map[string]interface{}) string {
	if interviewType == "profile" {
		// Адаптация для профильных вопросов
		if step == 1 && context["mentioned_skills"] != nil {
			if skills, ok := context["mentioned_skills"].([]interface{}); ok && len(skills) > 0 {
				return fmt.Sprintf("💻 Вы упомянули %v. Расскажите подробнее о вашем опыте с этими технологиями и оцените свой уровень по каждой (1-5).", skills)
			}
		}

		if step == 3 && context["experience_level"] != nil {
			level := context["experience_level"].(string)
			switch level {
			case "junior":
				return "🌱 Как начинающий специалист, какие проекты вас больше всего привлекают для получения опыта?"
			case "senior":
				return "🚀 С вашим опытом, какие сложные и интересные задачи вы готовы решать?"
			}
		}
	} else if interviewType == "task" {
		// Адаптация для вопросов о задачах
		if step == 2 && context["task_complexity"] != nil {
			complexity := context["task_complexity"].(string)
			switch complexity {
			case "simple":
				return "🛠️ Для простой задачи укажите базовые навыки, которые нужны исполнителю (например: HTML/CSS - 2/5, базовый JavaScript - 1/5)"
			case "complex":
				return "🔧 Для сложной задачи детально опишите требования к навыкам и опыту (например: React - 4/5, Node.js - 3/5, опыт с API)"
			}
		}

		if step == 3 && context["mentioned_technologies"] != nil {
			if techs, ok := context["mentioned_technologies"].([]interface{}); ok && len(techs) > 0 {
				return fmt.Sprintf("💰 Учитывая использование %v, какой бюджет подходит для данной задачи? (укажите сумму в рублях)", techs)
			}
		}
	}

	return question
}
