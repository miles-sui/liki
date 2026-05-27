// Package reports implements the Reports application service — LLM interpretation
// with SSE streaming generation, persistence, and public sharing.
package reports

import (
	"context"

	"github.com/25types/25types/internal/app/domain"
)

// Repository is the persistence port for reports.
type Repository interface {
	Create(ctx context.Context, report *domain.Report) (int64, error)
	FindByID(ctx context.Context, id int64, userID int64) (*domain.Report, error)
	ListByUser(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error)
	SoftDelete(ctx context.Context, id int64, userID int64) (bool, error)
	CreateShare(ctx context.Context, reportID int64, token string) error
	FindShareByToken(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error)
	RevokeShare(ctx context.Context, reportID int64) error
}

// Chunk is a single piece of streamed LLM output.
type Chunk struct {
	Text  string
	Error error
	Done  bool
}

// Streamer streams LLM completions.
type Streamer interface {
	Stream(ctx context.Context, systemPrompt string, messages []Message) (<-chan Chunk, error)
}

// Message is a conversation message.
type Message struct {
	Role    string
	Content string
}

// TemplateResolver loads scene-specific prompt templates.
type TemplateResolver interface {
	Get(scene, subScene, locale string) *Template
}

// Template holds a parsed prompt template.
type Template struct {
	Role        string
	InputTmpl   string
	OutputGuide string
}

// BuildMessages constructs the system prompt and message list.
func (t *Template) BuildMessages(input string) (system string, messages []Message) {
	if t == nil {
		return "", []Message{{Role: "user", Content: input}}
	}
	content := input
	if t.OutputGuide != "" {
		content += "\n\n" + t.OutputGuide
	}
	return t.Role, []Message{{Role: "user", Content: content}}
}
