package extractor

const ProfilePrompt = `Роль: Эксперт-аналитик по HR и профессиональным навыкам

Задача: Проанализируй ответ пользователя и извлеки:
1. Навыки (технические и софт-скиллы)
2. Уровень каждого навыка (1-5)
3. Опыт работы
4. Интересы и цели
5. Поведенческие особенности

Ответ пользователя: "{{user_answer}}"

Верни результат в JSON формате:
{
  "skills": [{"name": "Python", "level": 3, "confidence": 0.8}],
  "experience": [...],
  "interests": [...],
  "soft_skills": [...],
  "goals": [...]
}`

const TaskPrompt = `Роль: Эксперт по анализу задач

Задача: Проанализируй описание задачи и извлеки:
1. Название задачи
2. Описание
3. Требуемые навыки с минимальным уровнем
4. Бюджет
5. Сроки

Описание задачи: "{{task_description}}"

Верни результат в JSON формате:
{
  "title": "Название задачи",
  "description": "Подробное описание",
  "required_skills": {"Python": 3, "Machine Learning": 2},
  "budget": 50000,
  "deadline_days": 14
}`
