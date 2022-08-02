package main

import (
	"encoding/json"
	"os"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"sigs.k8s.io/yaml"
)

var validate = validator.New()

func loadConf(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	data = []byte(os.ExpandEnv(string(data)))
	js, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := json.Unmarshal(js, c); err != nil {
		return nil, err
	}
	if err := defaults.Set(c); err != nil {
		return nil, err
	}
	if err := validate.Struct(c); err != nil {
		return nil, err
	}
	return c, nil
}

type Config struct {
	Debug   bool          `json:"debug"`
	Slack   SlackConfig   `json:"slack" validate:"required"`
	GitHub  GitHubConfig  `json:"github" validate:"required"`
	Release ReleaseConfig `json:"release" validate:"required"`
}

// for each external service

type SlackConfig struct {
	BotToken string `json:"botToken" validate:"required"`
	AppToken string `json:"appToken" validate:"required"`
}

type GitHubConfig struct {
	Username    string `json:"username" validate:"required"`
	AccessToken string `json:"accessToken" validate:"required"`
}

// for each subcommand

type ReleaseConfig struct {
	Targets    []string `json:"targets" validate:"required"`
	BaseBranch string   `json:"baseBranch" default:"main"`
}
