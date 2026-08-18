package ui

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Step struct {
	Name      string   `yaml:"name"`
	Dir       string   `yaml:"dir"`
	DependsOn []string `yaml:"depends_on"`
}

type Pipeline struct {
	Steps []Step `yaml:"steps"`
}

func LoadSteps(filePath string) (Pipeline, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Pipeline{}, fmt.Errorf("read file: %w", err)
	}

	var cfg Pipeline
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return Pipeline{}, fmt.Errorf("yaml error: %w", err)
	}

	return Pipeline{cfg.Steps}, nil
}
