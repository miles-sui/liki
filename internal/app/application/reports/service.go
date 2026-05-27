package reports

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/25types/25types/internal/app/domain"
)

// Service orchestrates report generation, persistence, and sharing.
type Service struct {
	repo Repository
	llm  Streamer
	tmpl TemplateResolver
}

// NewService creates a new reports application service.
func NewService(repo Repository, llm Streamer, tmpl TemplateResolver) *Service {
	return &Service{repo: repo, llm: llm, tmpl: tmpl}
}

// Generate starts a streaming report generation. It validates the request,
// selects and renders the prompt template, calls the LLM, and saves the
// completed report asynchronously.
func (s *Service) Generate(ctx context.Context, req domain.CreateReportRequest, userID int64) (<-chan Chunk, error) {
	if !domain.ValidScene(string(req.Scene)) {
		return nil, domain.ErrInvalidScene
	}
	if len(req.EngineData) == 0 {
		return nil, domain.ErrEngineDataReq
	}
	if req.Locale == "" {
		req.Locale = "zh-CN"
	}

	tmpl := s.tmpl.Get(string(req.Scene), req.SubScene, req.Locale)
	rendered := renderTemplate(tmpl, req.EngineData, req.Question)
	sysPrompt, msgs := tmpl.BuildMessages(rendered)

	appMsgs := make([]Message, len(msgs))
	for i, m := range msgs {
		appMsgs[i] = Message{Role: m.Role, Content: m.Content}
	}

	chunks, err := s.llm.Stream(ctx, sysPrompt, appMsgs)
	if err != nil {
		return nil, fmt.Errorf("reports: llm stream: %w", err)
	}

	out := make(chan Chunk, 64)
	go s.collectAndSave(ctx, chunks, out, req, userID)
	return out, nil
}

// collectAndSave reads from the LLM chunk channel, forwards every chunk to the
// caller via out, and persists the completed report once the stream ends.
func (s *Service) collectAndSave(ctx context.Context, in <-chan Chunk, out chan Chunk, req domain.CreateReportRequest, userID int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("reports: panic in collectAndSave: %v\n%s", r, debug.Stack())
			// Best-effort: send error before closing; if channel is full we skip.
			select {
			case out <- Chunk{Error: fmt.Errorf("reports: internal error"), Done: true}:
			default:
			}
		}
		close(out)
	}()

	var fullContent strings.Builder
	for c := range in {
		if c.Error != nil {
			out <- c
			return
		}
		if c.Done {
			report := &domain.Report{
				UserID:     userID,
				Scene:      req.Scene,
				SubScene:   req.SubScene,
				Question:   req.Question,
				EngineData: req.EngineData,
				Content:    fullContent.String(),
				Locale:     req.Locale,
			}
			id, err := s.repo.Create(ctx, report)
			if err != nil {
				out <- Chunk{Error: fmt.Errorf("reports: save: %w", err), Done: true}
				return
			}
			out <- Chunk{Done: true, Text: fmt.Sprintf("%d", id)}
			return
		}
		fullContent.WriteString(c.Text)
		out <- c
	}
}

// ListHistory returns a user's report history, sorted by newest first.
func (s *Service) ListHistory(ctx context.Context, userID int64, scene string, limit, offset int) (*domain.ReportList, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, total, err := s.repo.ListByUser(ctx, userID, scene, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("reports: list: %w", err)
	}
	return &domain.ReportList{Items: items, Total: total}, nil
}

// GetDetail returns a single report by ID, if it belongs to the user.
func (s *Service) GetDetail(ctx context.Context, id, userID int64) (*domain.Report, error) {
	report, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("reports: get: %w", err)
	}
	return report, nil
}

// Delete soft-deletes a report owned by the user.
func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	deleted, err := s.repo.SoftDelete(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("reports: delete: %w", err)
	}
	if !deleted {
		return domain.ErrReportNotFound
	}
	return nil
}

// CreateShare creates a public sharing token for a report.
func (s *Service) CreateShare(ctx context.Context, reportID, userID int64) (*domain.ReportShare, error) {
	if _, err := s.repo.FindByID(ctx, reportID, userID); err != nil {
		return nil, domain.ErrReportNotFound
	}
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("reports: generate token: %w", err)
	}
	if err := s.repo.CreateShare(ctx, reportID, token); err != nil {
		return nil, fmt.Errorf("reports: share: %w", err)
	}
	return &domain.ReportShare{Token: token, ReportID: reportID}, nil
}

// GetShared returns a publicly shared report by token.
func (s *Service) GetShared(ctx context.Context, token string) (*domain.Report, error) {
	_, report, err := s.repo.FindShareByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrReportNotFound
	}
	return report, nil
}

// renderTemplate replaces {key} placeholders in the template with values from engine_data.
func renderTemplate(tmpl *Template, engineData json.RawMessage, question string) string {
	if tmpl == nil {
		return question
	}
	input := tmpl.InputTmpl

	var data map[string]any
	if err := json.Unmarshal(engineData, &data); err != nil {
		if question != "" {
			input = strings.ReplaceAll(input, "{question}", question)
		}
		return input
	}

	for k, v := range data {
		placeholder := "{" + k + "}"
		input = strings.ReplaceAll(input, placeholder, fmt.Sprintf("%v", v))
	}
	if question != "" {
		input = strings.ReplaceAll(input, "{question}", question)
	}
	return input
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
