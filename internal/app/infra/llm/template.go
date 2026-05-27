package llm

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/25types/25types/internal/app/application/reports"
	"gopkg.in/yaml.v3"
)

var _ reports.TemplateResolver = (*TemplateSet)(nil)

//go:embed data/*.yaml
var templateFS embed.FS

// TemplateSet loads and indexes templates by key.
type TemplateSet struct {
	templates map[string]*reports.Template
}

// LoadTemplates reads all YAML files from the embedded data.
func LoadTemplates() (*TemplateSet, error) {
	ts := &TemplateSet{templates: make(map[string]*reports.Template)}
	err := fs.WalkDir(templateFS, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return err
		}
		t, err := loadTemplateFS(path)
		if err != nil {
			return fmt.Errorf("llm: load %s: %w", d.Name(), err)
		}
		key := strings.TrimSuffix(d.Name(), ".yaml")
		ts.templates[key] = t
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("llm: load templates: %w", err)
	}
	return ts, nil
}

func loadTemplateFS(path string) (*reports.Template, error) {
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Role          string `yaml:"role"`
		InputTemplate string `yaml:"input_template"`
		OutputGuide   string `yaml:"output_guide"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &reports.Template{
		Role:        raw.Role,
		InputTmpl:   raw.InputTemplate,
		OutputGuide: raw.OutputGuide,
	}, nil
}

// Get returns the template for a given scene + locale key.
func (ts *TemplateSet) Get(scene, subScene, locale string) *reports.Template {
	if subScene != "" {
		key := fmt.Sprintf("%s_%s_%s", scene, subScene, locale)
		if t := ts.templates[key]; t != nil {
			return t
		}
	}
	key := fmt.Sprintf("%s_%s", scene, locale)
	return ts.templates[key]
}

// MustLoadTemplates is a convenience wrapper that panics on error, for use at startup.
func MustLoadTemplates() *TemplateSet {
	ts, err := LoadTemplates()
	if err != nil {
		panic(err)
	}
	if len(ts.templates) == 0 {
		panic("llm: no templates found in embedded data")
	}
	return ts
}
