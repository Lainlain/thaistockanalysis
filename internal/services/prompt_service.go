package services

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"regexp"
	"time"
)

// PromptService is responsible for generating dynamic, human-like prompts from a JSON file.
type PromptService struct {
	highlightTemplates map[string][]string
}

// NewPromptService creates a new instance of PromptService and loads the templates from JSON.
func NewPromptService(jsonPath string) (*PromptService, error) {
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var templates map[string][]string
	if err := json.Unmarshal(data, &templates); err != nil {
		return nil, err
	}

	return &PromptService{
		highlightTemplates: templates,
	}, nil
}

// GenerateHighlightNarrative takes a raw string of numbers, identifies the last digit of the first number,
// and returns a random, human-like narrative sentence from the loaded templates.
func (s *PromptService) GenerateHighlightNarrative(rawHighlights string) string {
	// Use a regular expression to find the first number in the input string.
	re := regexp.MustCompile(`[+-]?(\d+)`)
	match := re.FindStringSubmatch(rawHighlights)

	if len(match) < 2 {
		return "No specific market-moving highlights were noted in this session."
	}

	// Get the first number found.
	firstNumberStr := match[1]

	// Get the last character of the number string, which represents the key in our JSON.
	lastKey := string(firstNumberStr[len(firstNumberStr)-1])

	// Look up the available sentences for this key.
	sentences, ok := s.highlightTemplates[lastKey]
	if !ok || len(sentences) == 0 {
		return "General market activity was observed without a distinct focus."
	}

	// Create a new random generator with a new source
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(sentences))

	return sentences[randomIndex]
}
