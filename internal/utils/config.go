package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Scraper struct {
		Timeout  int `yaml:"timeout"`
		Retries  int `yaml:"retries"`
		Delay    int `yaml:"delay"`
		MaxPages int `yaml:"maxPages"`
		Browser  struct {
			Headless bool `yaml:"headless"`
			Debug    bool `yaml:"debug"`
		} `yaml:"browser"`
	} `yaml:"scraper"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
