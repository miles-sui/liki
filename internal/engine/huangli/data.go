package huangli

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed data/jianchu.yaml
var jianchuYAML []byte

var jianChuCfg jianchuConfig

func init() {
	if err := loadJianchu(); err != nil {
		log.Fatalf("huangli: load jianchu: %v", err)
	}
}

func loadJianchu() error {
	// Parse into anonymous struct — yaml.v3 skips map values with named types.
	var raw struct {
		Jianchu struct {
			Sequence   []string            `yaml:"sequence"`
			Suitable   map[string][]string `yaml:"suitable"`
			Forbidden  map[string][]string `yaml:"forbidden"`
			EventRules map[string]struct {
				Label     string   `yaml:"label"`
				Suitable  []string `yaml:"suitable"`
				Forbidden []string `yaml:"forbidden"`
			} `yaml:"event_rules"`
			ShenSha map[string]map[string][]string `yaml:"shensha"`
		} `yaml:"jianchu"`
	}
	if err := yaml.Unmarshal(jianchuYAML, &raw); err != nil {
		return err
	}

	cfg := raw.Jianchu
	c := &jianChuCfg
	c.Sequence = cfg.Sequence
	c.Suitable = cfg.Suitable
	c.Forbidden = cfg.Forbidden

	c.EventRules = make(map[string]eventRule, len(cfg.EventRules))
	for k, v := range cfg.EventRules {
		c.EventRules[k] = eventRule{Label: v.Label, Suitable: v.Suitable, Forbidden: v.Forbidden}
	}

	c.ShenSha = make(map[string]map[string][]string, len(cfg.ShenSha))
	for k, v := range cfg.ShenSha {
		c.ShenSha[k] = v
	}
	return nil
}
