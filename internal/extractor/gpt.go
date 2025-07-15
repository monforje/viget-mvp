package extractor

import (
	"strings"
	"viget-mvp/pkg/gpt"
)

type Extractor struct {
	gptClient *gpt.Client
}

func NewExtractor(gptClient *gpt.Client) *Extractor {
	return &Extractor{gptClient: gptClient}
}

func (e *Extractor) ExtractProfile(userAnswer string) (string, error) {
	prompt := strings.Replace(ProfilePrompt, "{{user_answer}}", userAnswer, 1)
	return e.gptClient.SendRequest(prompt)
}

func (e *Extractor) ExtractTask(taskDescription string) (string, error) {
	prompt := strings.Replace(TaskPrompt, "{{task_description}}", taskDescription, 1)
	return e.gptClient.SendRequest(prompt)
}
