package vibot

import "fmt"

type QuestionBank struct {
	profileQuestions []string
	taskQuestions    []string
}

func NewQuestionBank() *QuestionBank {
	return &QuestionBank{
		profileQuestions: []string{
			"Привет! Как вас зовут?",
			"Расскажите о своем опыте в IT. Какими технологиями владеете?",
			"Какой у вас уровень в программировании? Есть ли любимые языки?",
			"Что вас интересует в работе? Какие проекты хотели бы делать?",
			"Какие у вас сильные стороны в работе с людьми?",
			"Какие цели хотите достичь в ближайшее время?",
			"Есть ли опыт удаленной работы или фриланса?",
		},
		taskQuestions: []string{
			"Как называется ваша задача?",
			"Опишите подробно, что нужно сделать",
			"Какие навыки нужны исполнителю? Какой уровень?",
			"Какой бюджет вы готовы выделить?",
			"В какие сроки нужно выполнить задачу?",
			"Есть ли особые требования к исполнителю?",
		},
	}
}

func (q *QuestionBank) GetQuestion(interviewType string, step int, context map[string]interface{}) string {
	var questions []string

	switch interviewType {
	case "profile":
		questions = q.profileQuestions
	case "task":
		questions = q.taskQuestions
	default:
		return "Неизвестный тип интервью"
	}

	if step >= len(questions) {
		return "Интервью завершено"
	}

	question := questions[step]

	// Адаптируем вопрос на основе контекста
	if len(context) > 0 {
		question = q.adaptQuestion(question, step, context)
	}

	return question
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

func (q *QuestionBank) adaptQuestion(question string, step int, context map[string]interface{}) string {
	// Простая адаптация вопросов на основе контекста
	if step == 2 && context["mentioned_skills"] != nil {
		if skills, ok := context["mentioned_skills"].([]interface{}); ok && len(skills) > 0 {
			return fmt.Sprintf("Вы упомянули %v. Расскажите подробнее о вашем уровне в этих технологиях?", skills)
		}
	}

	if step == 3 && context["experience_level"] != nil {
		level := context["experience_level"].(string)
		switch level {
		case "junior":
			return "Как junior разработчик, какие проекты вас больше всего привлекают для получения опыта?"
		case "senior":
			return "С вашим опытом, какие сложные задачи вы готовы решать?"
		}
	}

	return question
}
