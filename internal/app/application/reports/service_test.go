package reports_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/25types/25types/internal/app/application/reports"
	"github.com/25types/25types/internal/app/domain"
)

// --- mocks ---

type mockRepo struct {
	createFn      func(ctx context.Context, report *domain.Report) (int64, error)
	findByIDFn    func(ctx context.Context, id, userID int64) (*domain.Report, error)
	listFn        func(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error)
	softDeleteFn  func(ctx context.Context, id, userID int64) (bool, error)
	createShareFn func(ctx context.Context, reportID int64, token string) error
	findShareFn   func(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error)
	revokeShareFn func(ctx context.Context, reportID int64) error
}

func (m *mockRepo) Create(ctx context.Context, r *domain.Report) (int64, error) {
	return m.createFn(ctx, r)
}
func (m *mockRepo) FindByID(ctx context.Context, id, userID int64) (*domain.Report, error) {
	return m.findByIDFn(ctx, id, userID)
}
func (m *mockRepo) ListByUser(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error) {
	return m.listFn(ctx, userID, scene, limit, offset)
}
func (m *mockRepo) SoftDelete(ctx context.Context, id, userID int64) (bool, error) {
	return m.softDeleteFn(ctx, id, userID)
}
func (m *mockRepo) CreateShare(ctx context.Context, reportID int64, token string) error {
	return m.createShareFn(ctx, reportID, token)
}
func (m *mockRepo) FindShareByToken(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error) {
	return m.findShareFn(ctx, token)
}
func (m *mockRepo) RevokeShare(ctx context.Context, reportID int64) error {
	return m.revokeShareFn(ctx, reportID)
}

type mockTmpl struct{}

func (m *mockTmpl) Get(scene, subScene, locale string) *reports.Template {
	if scene == "mingli" {
		return &reports.Template{
			Role:        "你是一个八字命理专家",
			InputTmpl:   "分析这个八字: {chart}",
			OutputGuide: "请用中文回答",
		}
	}
	return nil
}

// --- tests ---

func TestNewService(t *testing.T) {
	repo := &mockRepo{}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestGenerate_InvalidScene(t *testing.T) {
	svc := reports.NewService(&mockRepo{}, nil, &mockTmpl{})
	_, err := svc.Generate(context.Background(), domain.CreateReportRequest{
		Scene:      "invalid",
		EngineData: json.RawMessage(`{}`),
	}, 1)
	if !errors.Is(err, domain.ErrInvalidScene) {
		t.Fatalf("expected ErrInvalidScene, got %v", err)
	}
}

func TestGenerate_MissingEngineData(t *testing.T) {
	svc := reports.NewService(&mockRepo{}, nil, &mockTmpl{})
	_, err := svc.Generate(context.Background(), domain.CreateReportRequest{
		Scene: domain.SceneMingli,
	}, 1)
	if !errors.Is(err, domain.ErrEngineDataReq) {
		t.Fatalf("expected ErrEngineDataReq, got %v", err)
	}
}

func TestListHistory(t *testing.T) {
	repo := &mockRepo{
		listFn: func(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error) {
			return []domain.ReportItem{{ID: 1, Scene: "mingli"}}, 1, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	result, err := svc.ListHistory(context.Background(), 1, "", 10, 0)
	if err != nil {
		t.Fatalf("ListHistory: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Total != 1 {
		t.Errorf("expected total=1, got %d", result.Total)
	}
}

func TestListHistory_DefaultLimit(t *testing.T) {
	repo := &mockRepo{
		listFn: func(ctx context.Context, userID int64, scene string, limit, offset int) ([]domain.ReportItem, int, error) {
			if limit != 20 {
				t.Errorf("expected default limit=20, got %d", limit)
			}
			return nil, 0, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	svc.ListHistory(context.Background(), 1, "", 0, 0)
}

func TestGetDetail(t *testing.T) {
	expected := &domain.Report{ID: 42, UserID: 1, Content: "hello"}
	repo := &mockRepo{
		findByIDFn: func(ctx context.Context, id, userID int64) (*domain.Report, error) {
			return expected, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	r, err := svc.GetDetail(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if r.Content != "hello" {
		t.Errorf("expected content='hello', got %q", r.Content)
	}
}

func TestDelete(t *testing.T) {
	repo := &mockRepo{
		softDeleteFn: func(ctx context.Context, id, userID int64) (bool, error) {
			return true, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	if err := svc.Delete(context.Background(), 42, 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockRepo{
		softDeleteFn: func(ctx context.Context, id, userID int64) (bool, error) {
			return false, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	err := svc.Delete(context.Background(), 99, 1)
	if !errors.Is(err, domain.ErrReportNotFound) {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}
}

func TestCreateShare(t *testing.T) {
	repo := &mockRepo{
		findByIDFn: func(ctx context.Context, id, userID int64) (*domain.Report, error) {
			return &domain.Report{ID: 1, UserID: 1}, nil
		},
		createShareFn: func(ctx context.Context, reportID int64, token string) error {
			return nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	share, err := svc.CreateShare(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}
	if share.Token == "" {
		t.Error("expected non-empty token")
	}
	if share.ReportID != 1 {
		t.Errorf("expected reportID=1, got %d", share.ReportID)
	}
}

func TestCreateShare_ReportNotFound(t *testing.T) {
	repo := &mockRepo{
		findByIDFn: func(ctx context.Context, id, userID int64) (*domain.Report, error) {
			return nil, domain.ErrReportNotFound
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	_, err := svc.CreateShare(context.Background(), 99, 1)
	if !errors.Is(err, domain.ErrReportNotFound) {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}
}

func TestGetShared(t *testing.T) {
	expected := &domain.Report{ID: 1, Content: "shared content"}
	repo := &mockRepo{
		findShareFn: func(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error) {
			return &domain.ReportShare{Token: token}, expected, nil
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	r, err := svc.GetShared(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetShared: %v", err)
	}
	if r.Content != "shared content" {
		t.Errorf("expected 'shared content', got %q", r.Content)
	}
}

func TestGetShared_NotFound(t *testing.T) {
	repo := &mockRepo{
		findShareFn: func(ctx context.Context, token string) (*domain.ReportShare, *domain.Report, error) {
			return nil, nil, domain.ErrReportNotFound
		},
	}
	svc := reports.NewService(repo, nil, &mockTmpl{})
	_, err := svc.GetShared(context.Background(), "bad-token")
	if !errors.Is(err, domain.ErrReportNotFound) {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}
}
