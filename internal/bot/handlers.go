// internal/bot/handlers.go (обновленная версия с исправлениями)
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
	case strings.HasPrefix(text, "/cancel"):
		h.handleCancel(userID)
	default:
		// Если пользователь в процессе интервью
		if h.interviewer.IsInInterview(userID) {
			h.handleInterviewAnswer(userID, text)
		} else {
			h.sendMessage(userID, "❓ Не понимаю команду. Используйте /help для справки.")
		}
	}
}

func (h *Handler) handleStart(userID int64) {
	profile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))

	var msg string
	if profile == nil {
		msg = `👋 Добро пожаловать в Viget!

🤖 Я помогу создать ваш цифровой профиль и найти подходящие задачи.

Для начала пройдите интервью: /interview`
	} else {
		msg = fmt.Sprintf(`👋 С возвращением, %s!

✅ Ваш профиль готов. Вы можете:
• Посмотреть профиль: /profile
• Найти задачи: /tasks
• Создать задачу: /create_task`, profile.Name)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Создать профиль", "interview"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Найти задачи", "tasks"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Мой профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать задачу", "create_task"),
		),
	)

	h.sendMessageWithKeyboard(userID, msg, keyboard)
}

func (h *Handler) handleProfile(userID int64) {
	profile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))

	if profile == nil {
		h.sendMessage(userID, "❌ У вас еще нет профиля.\n\n🚀 Пройдите интервью: /interview")
		return
	}

	msg := fmt.Sprintf(`👤 **Ваш профиль:**

🏷️ **Имя:** %s
🛠️ **Навыки:** %s
💡 **Интересы:** %s
🎯 **Цели:** %s

📅 Создан: %s
🔄 Обновлен: %s`,
		profile.Name,
		h.formatSkills(profile.Skills),
		strings.Join(profile.Interests, ", "),
		strings.Join(profile.Goals, ", "),
		profile.CreatedAt.Format("02.01.2006"),
		profile.UpdatedAt.Format("02.01.2006"))

	h.sendMessage(userID, msg)
}

func (h *Handler) handleInterview(userID int64) {
	// Проверяем, не находится ли пользователь уже в интервью
	if h.interviewer.IsInInterview(userID) {
		currentType := h.interviewer.GetInterviewType(userID)
		var typeMsg string
		if currentType == "profile" {
			typeMsg = "создания профиля"
		} else {
			typeMsg = "создания задачи"
		}
		h.sendMessage(userID, fmt.Sprintf("⚠️ Вы уже проходите интервью для %s.\n\nИспользуйте /cancel для отмены.", typeMsg))
		return
	}

	err := h.interviewer.StartInterview(userID, "profile")
	if err != nil {
		h.sendMessage(userID, "❌ Ошибка запуска интервью. Попробуйте позже.")
		return
	}

	question := h.interviewer.GetCurrentQuestion(userID)
	h.sendMessage(userID, question+"\n\n💡 Используйте /cancel для отмены интервью")
}

func (h *Handler) handleCreateTask(userID int64) {
	// Проверяем, не находится ли пользователь уже в интервью
	if h.interviewer.IsInInterview(userID) {
		currentType := h.interviewer.GetInterviewType(userID)
		var typeMsg string
		if currentType == "profile" {
			typeMsg = "создания профиля"
		} else {
			typeMsg = "создания задачи"
		}
		h.sendMessage(userID, fmt.Sprintf("⚠️ Вы уже проходите интервью для %s.\n\nИспользуйте /cancel для отмены.", typeMsg))
		return
	}

	err := h.interviewer.StartInterview(userID, "task")
	if err != nil {
		h.sendMessage(userID, "❌ Ошибка запуска создания задачи. Попробуйте позже.")
		return
	}

	question := h.interviewer.GetCurrentQuestion(userID)
	h.sendMessage(userID, question+"\n\n💡 Используйте /cancel для отмены")
}

func (h *Handler) handleCancel(userID int64) {
	if !h.interviewer.IsInInterview(userID) {
		h.sendMessage(userID, "❌ Вы не проходите интервью.")
		return
	}

	// Получаем тип интервью перед удалением сессии
	interviewType := h.interviewer.GetInterviewType(userID)

	// Удаляем сессию (добавим этот метод в interviewer)
	h.interviewer.CancelInterview(userID)

	var typeMsg string
	if interviewType == "profile" {
		typeMsg = "создания профиля"
	} else {
		typeMsg = "создания задачи"
	}

	h.sendMessage(userID, fmt.Sprintf("❌ Интервью для %s отменено.\n\n🔄 Используйте /start для возврата в главное меню.", typeMsg))
}

func (h *Handler) handleInterviewAnswer(userID int64, answer string) {
	interviewType := h.interviewer.GetInterviewType(userID)

	nextQuestion, finished, err := h.interviewer.ProcessAnswer(userID, answer)
	if err != nil {
		h.sendMessage(userID, "❌ Ошибка обработки ответа. Попробуйте еще раз.")
		return
	}

	if finished {
		switch interviewType {
		case "profile":
			// Извлекаем профиль и сохраняем
			profile, err := h.interviewer.ExtractProfile(userID)
			if err != nil {
				h.sendMessage(userID, "❌ Ошибка создания профиля. Попробуйте позже.")
				return
			}

			h.storage.SaveUserProfile(profile)
			h.sendMessage(userID, "✅ Интервью завершено! Ваш профиль создан.\n\n🎯 Теперь вы можете искать задачи: /tasks")

		case "task":
			// Извлекаем задачу и сохраняем
			task, err := h.interviewer.ExtractTask(userID)
			if err != nil {
				h.sendMessage(userID, "❌ Ошибка создания задачи. Попробуйте позже.")
				return
			}

			h.storage.SaveTask(task)

			msg := fmt.Sprintf(`✅ Задача успешно создана!

📋 **%s**
💰 Бюджет: %d ₽
⏰ Дедлайн: %s

🎯 Ваша задача добавлена в систему и скоро появится у подходящих исполнителей.`,
				task.Title,
				task.Budget,
				task.Deadline.Format("02.01.2006"))

			h.sendMessage(userID, msg)
		}
	} else {
		h.sendMessage(userID, nextQuestion+"\n\n💡 Используйте /cancel для отмены интервью")
	}
}

func (h *Handler) handleTasks(userID int64) {
	userProfile := h.storage.GetUserProfile(strconv.FormatInt(userID, 10))
	if userProfile == nil {
		h.sendMessage(userID, "❌ Сначала создайте профиль: /interview")
		return
	}

	tasks := h.storage.GetAvailableTasks()
	if len(tasks) == 0 {
		h.sendMessage(userID, "😔 Пока нет доступных задач. Проверьте позже!\n\n➕ Или создайте свою задачу: /create_task")
		return
	}

	matches := h.matcher.FindMatchingTasks(userProfile, tasks)

	if len(matches) == 0 {
		h.sendMessage(userID, "😕 Не найдено подходящих задач. Попробуйте обновить профиль: /interview")
		return
	}

	msg := "🎯 **Рекомендованные задачи:**\n\n"
	for i, match := range matches {
		if i >= 5 { // Показываем только топ-5
			break
		}

		task := h.storage.GetTask(match.TaskID)
		msg += fmt.Sprintf(`📋 **%s**
💰 %d ₽
🎯 Совпадение: %.0f%%
⏰ До %s

`, task.Title, task.Budget, match.Score*100, task.Deadline.Format("02.01"))
	}

	msg += "\n💡 Для получения полной информации о задаче свяжитесь с @monforje"

	h.sendMessage(userID, msg)
}

func (h *Handler) handleHelp(userID int64) {
	msg := `🤖 **Viget** - помощник для поиска задач и исполнителей

**Команды:**
/start - Главное меню
/profile - Ваш профиль  
/interview - Пройти интервью для создания профиля
/tasks - Найти подходящие задачи
/create_task - Создать задачу для исполнителей
/cancel - Отменить текущее интервью
/help - Эта справка

**Как это работает:**
1️⃣ Пройдите интервью для создания профиля
2️⃣ Получайте персональные рекомендации задач
3️⃣ Или создавайте свои задачи для других

**Поддержка:** @monforje`

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
