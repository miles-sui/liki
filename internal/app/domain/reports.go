// Package domain defines the Reports aggregate — LLM interpretation reports
// with SSE streaming generation, persistence, and public sharing.
package domain

import (
	"encoding/json"
	"time"
)

// Scene identifies the top-level product domain for a report.
type Scene string

const (
	SceneMingli         Scene = "mingli"
	SceneQiming       Scene = "qiming"
	SceneHuangli      Scene = "huangli"
	SceneRelationship Scene = "relationship"
	SceneCareer       Scene = "career"
	SceneGeneral      Scene = "general"
)

// ValidScene reports whether s is a known Scene.
func ValidScene(s string) bool {
	switch Scene(s) {
	case SceneMingli, SceneQiming, SceneHuangli, SceneRelationship, SceneCareer, SceneGeneral:
		return true
	}
	return false
}

// Report is the aggregate root — an LLM-generated interpretation.
type Report struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	Scene      Scene           `json:"scene"`
	SubScene   string          `json:"sub_scene,omitempty"`
	Question   string          `json:"question,omitempty"`
	EngineData json.RawMessage `json:"engine_data"`
	Content    string          `json:"content"`
	Locale     string          `json:"locale"`
	CreatedAt  time.Time       `json:"created_at"`
}

// CreateReportRequest is the input for starting a new report generation.
type CreateReportRequest struct {
	Scene      Scene           `json:"scene"`
	SubScene   string          `json:"sub_scene,omitempty"`
	Question   string          `json:"question,omitempty"`
	EngineData json.RawMessage `json:"engine_data"`
	Locale     string          `json:"locale"`
}

// ReportItem is a summary row for the history list (content omitted).
type ReportItem struct {
	ID        int64     `json:"id"`
	Scene     Scene     `json:"scene"`
	SubScene  string    `json:"sub_scene,omitempty"`
	Question  string    `json:"question,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ReportList is a paginated collection of ReportItem summaries.
type ReportList struct {
	Items []ReportItem `json:"items"`
	Total int          `json:"total"`
}

// ReportShare is a public sharing token for a report.
type ReportShare struct {
	Token     string     `json:"token"`
	ReportID  int64      `json:"report_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
