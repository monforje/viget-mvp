package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"viget-mvp/internal/matcher"
	"viget-mvp/internal/models"
	"viget-mvp/internal/profile"
	"viget-mvp/internal/vibot"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	storage     *profile.InMemoryStorage
	interviewer *vibot.Interviewer
	matcher     *matcher.Matcher
}

func NewHandler(bot *tgbotapi.BotAPI, storage *profile.InMemoryStorage,
	interviewer *vibot.Interviewer, matcher *matcher.Matcher) *Handler {
	return &Handler{
		bot:         bot,
		storage:     storage,
		interviewer: interviewer,
		matcher:     matcher,
	}
}

func (h *Handler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			h.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			h.handleCallback(update.CallbackQuery)
		}
	}
}

func (h *Handler) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	text := message.Text

	switch {
	case strings.HasPrefix(text, "/start"):
		h.handleStart(userID)
	case strings.HasPrefix(text, "/profile"):
		h.handleProfile(userID)
	case strings.HasPrefix(text, "/interview"):
		h.handleInterview(userID)
	case strings.HasPrefix(text, "/tasks"):
		h.handleTasks(userID)
	case strings.HasPrefix(text, "/create_task"):
		h.handleCreateTask(userID)
	case strings.HasPrefix(text, "/help"):
		h.handleHelp(userID)
	default:
		// Если пользователь в процессе интервью
		if h.interviewer.IsInInterview(userID) {
			h.handleInterviewAnswer(userID, text)
		} else {
			h.sendMessage(userID, "Не понимаю команду. Используйте /help для справки.")
		}
	}
}

func (h *Handler) handleStart(userID int64) {
	profile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))

	var msg string
	if profile == nil {
		msg = `👋 Добро пожаловать в Viget!

Я помогу создать ваш цифровой профиль и найти подходящие задачи.

Для начала пройдите интервью: /interview`
	} else {
		msg = fmt.Sprintf(`👋 С возвращением, %s!

Ваш профиль готов. Вы можете:
• Посмотреть профиль: /profile
• Найти задачи: /tasks
• Создать задачу: /create_task`, profile.Name)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Пройти интервью", "interview"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Найти задачи", "tasks"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать задачу", "create_task"),
		),
	)

	h.sendMessageWithKeyboard(userID, msg, keyboard)
}

func (h *Handler) handleProfile(userID int64) {
	profile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))

	if profile == nil {
		h.sendMessage(userID, "У вас еще нет профиля. Пройдите интервью: /interview")
		return
	}

	msg := fmt.Sprintf(`👤 Ваш профиль:

🏷️ Имя: %s
🛠️ Навыки: %s
💡 Интересы: %s
🎯 Цели: %s

Создан: %s
Обновлен: %s`,
		profile.Name,
		h.formatSkills(profile.Skills),
		strings.Join(profile.Interests, ", "),
		strings.Join(profile.Goals, ", "),
		profile.CreatedAt.Format("02.01.2006"),
		profile.UpdatedAt.Format("02.01.2006"))

	h.sendMessage(userID, msg)
}

func (h *Handler) handleInterview(userID int64) {
	err := h.interviewer.StartInterview(userID, "profile")
	if err != nil {
		h.sendMessage(userID, "Ошибка запуска интервью. Попробуйте позже.")
		return
	}

	question := h.interviewer.GetCurrentQuestion(userID)
	h.sendMessage(userID, question)
}

func (h *Handler) handleInterviewAnswer(userID int64, answer string) {
	nextQuestion, finished, err := h.interviewer.ProcessAnswer(userID, answer)
	if err != nil {
		h.sendMessage(userID, "Ошибка обработки ответа. Попробуйте еще раз.")
		return
	}

	if finished {
		// Извлекаем профиль и сохраняем
		profile, err := h.interviewer.ExtractProfile(userID)
		if err != nil {
			h.sendMessage(userID, "Ошибка создания профиля. Попробуйте позже.")
			return
		}

		h.storage.SaveUserProfile(profile)
		h.sendMessage(userID, "✅ Интервью завершено! Ваш профиль создан.\n\nТеперь вы можете искать задачи: /tasks")
	} else {
		h.sendMessage(userID, nextQuestion)
	}
}

func (h *Handler) handleTasks(userID int64) {
	userProfile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))
	if userProfile == nil {
		h.sendMessage(userID, "Сначала создайте профиль: /interview")
		return
	}

	tasks := h.storage.GetAvailableTasks()
	if len(tasks) == 0 {
		h.sendMessage(userID, "Пока нет доступных задач. Проверьте позже!")
		return
	}

	matches := h.matcher.FindMatchingTasks(userProfile, tasks)

	if len(matches) == 0 {
		h.sendMessage(userID, "Не найдено подходящих задач. Попробуйте обновить профиль.")
		return
	}

	msg := "🎯 Рекомендованные задачи:\n\n"
	for i, match := range matches {
		if i >= 5 { // Показываем только топ-5
			break
		}

		task := h.storage.GetTask(match.TaskID)
		msg += fmt.Sprintf("📋 %s\n💰 %d ₽\n🎯 Совпадение: %.0f%%\n\n",
			task.Title, task.Budget, match.Score*100)
	}

	h.sendMessage(userID, msg)
}

func (h *Handler) handleCreateTask(userID int64) {
	err := h.interviewer.StartInterview(userID, "task")
	if err != nil {
		h.sendMessage(userID, "Ошибка запуска создания задачи. Попробуйте позже.")
		return
	}

	question := h.interviewer.GetCurrentQuestion(userID)
	h.sendMessage(userID, "📝 Создание задачи\n\n"+question)
}

func (h *Handler) handleHelp(userID int64) {
	msg := `🤖 Viget - помощник для поиска задач и исполнителей

Команды:
/start - Главное меню
/profile - Ваш профиль
/interview - Пройти интервью
/tasks - Найти задачи
/create_task - Создать задачу
/help - Эта справка

Как это работает:
1️⃣ Пройдите интервью для создания профиля
2️⃣ Получайте персональные рекомендации задач
3️⃣ Или создавайте свои задачи для других

По вопросам: @monforje`

	h.sendMessage(userID, msg)
}

func (h *Handler) handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data

	// Отвечаем на callback
	h.bot.Send(tgbotapi.NewCallback(callback.ID, ""))

	switch data {
	case "interview":
		h.handleInterview(userID)
	case "tasks":
		h.handleTasks(userID)
	case "profile":
		h.handleProfile(userID)
	case "create_task":
		h.handleCreateTask(userID)
	}
}

func (h *Handler) sendMessage(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	h.bot.Send(msg)
}

func (h *Handler) sendMessageWithKeyboard(userID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) formatSkills(skills map[string]models.SkillLevel) string {
	if len(skills) == 0 {
		return "Не указаны"
	}

	var result []string
	for _, skill := range skills {
		verified := ""
		if skill.Verified {
			verified = " ✅"
		}
		result = append(result, fmt.Sprintf("%s (%d/5)%s", skill.Name, skill.Level, verified))
	}

	return strings.Join(result, ", ")
}
