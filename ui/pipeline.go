package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Step struct {
	Name      string   `yaml:"name"`
	Tool      string   `yaml: "tool"`
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

func contains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func CheckTools(steps Pipeline) error {
	var checked []string
	var missing []string

	for _, step := range steps.Steps {
		if step.Tool == "" || contains(checked, step.Tool) {
			continue
		}
		checked = append(checked, step.Tool)

		if _, err := exec.LookPath(step.Tool); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				missing = append(missing, step.Tool)
			} else {
				missing = append(missing, fmt.Sprintf("%s (%v)", step.Tool, err))
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required tools in PATH: %s", strings.Join(missing, "\n"))
	}

	return nil
}
