package configuration

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYamlFile(filepath string) (*Config, error) {

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
