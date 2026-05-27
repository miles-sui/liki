package questionnaire

import (
	"embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

// Meta holds questionnaire-level metadata.
type Meta struct {
	Version      string `yaml:"version"`
	Language     string `yaml:"language"`
	ResponseMode string `yaml:"response_mode"`
	Description  string `yaml:"description"`
}

// Option represents a single answer option within a question.
type Option struct {
	Text    string `yaml:"text" json:"text"`
	Element string `yaml:"element" json:"element"`
}

// Question represents a single triad forced-choice question.
type Question struct {
	ID      string   `yaml:"id" json:"qid"`
	Text    string   `yaml:"text" json:"text"`
	Options []Option `yaml:"options" json:"options"`
}

// Round holds one balanced block of 5 questions.
type Round struct {
	ID        string     `yaml:"id"`
	Questions []Question `yaml:"questions"`
}

// Questionnaire holds the parsed question bank in v4.0 rounds format.
type Questionnaire struct {
	Meta   Meta    `yaml:"meta"`
	Rounds []Round `yaml:"rounds"`
}

//go:embed data/en.yaml data/zh-CN.yaml
var assessmentFS embed.FS

var (
	mu            sync.RWMutex
	loaded        = map[string]*Questionnaire{}
	qidLists      = map[string][]string{}            // locale → ordered QID list
	questionIndex = map[string]map[string]Question{} // locale → QID → Question
)

// Load reads and parses the questionnaire for the given locale from embedded data.
// Only "en" and "zh-CN" are valid locales.
func Load(locale string) (*Questionnaire, error) {
	if locale != "en" && locale != "zh-CN" {
		return nil, fmt.Errorf("questionnaire: unsupported locale %q", locale)
	}
	mu.RLock()
	if q, ok := loaded[locale]; ok {
		mu.RUnlock()
		return q, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if q, ok := loaded[locale]; ok {
		return q, nil
	}

	b, err := assessmentFS.ReadFile("data/" + locale + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("questionnaire: cannot read embedded data for %s: %w", locale, err)
	}

	var q Questionnaire
	if err := yaml.Unmarshal(b, &q); err != nil {
		return nil, fmt.Errorf("questionnaire: parse error for %s: %w", locale, err)
	}

	loaded[locale] = &q

	var qids []string
	idx := map[string]Question{}
	for _, r := range q.Rounds {
		for _, vq := range r.Questions {
			qids = append(qids, vq.ID)
			idx[vq.ID] = vq
		}
	}
	qidLists[locale] = qids
	questionIndex[locale] = idx

	return loaded[locale], nil
}

// AllQIDs returns all question IDs in order for the default (en) locale.
func AllQIDs() []string {
	mu.RLock()
	defer mu.RUnlock()
	return qidLists["en"]
}

// GetQuestions returns the questions for the given QID list from the default locale.
func GetQuestions(q *Questionnaire, qids []string) []Question {
	mu.RLock()
	locale := ""
	for loc, qp := range loaded {
		if qp == q {
			locale = loc
			break
		}
	}
	idx := questionIndex[locale]
	mu.RUnlock()

	out := make([]Question, 0, len(qids))
	for _, id := range qids {
		if vq, ok := idx[id]; ok {
			out = append(out, vq)
		}
	}
	return out
}
