package huangli

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed data/jianchu.yaml
var jianchuYAML []byte

var JianChuConfig JianchuConfig

var defaultData = struct {
	JianChuConfig JianchuConfig
}{}

var defaultEngine = &defaultData

func init() {
	if err := loadJianchu(); err != nil {
		log.Printf("huangli: load jianchu: %v", err)
	}
	defaultData.JianChuConfig = JianChuConfig
}

func loadJianchu() error {
	var wrapper struct {
		Jianchu JianchuConfig `yaml:"jianchu"`
	}
	if err := yaml.Unmarshal(jianchuYAML, &wrapper); err != nil {
		return err
	}
	JianChuConfig = wrapper.Jianchu
	return nil
}
