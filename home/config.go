package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	ListenKcp   string `json:"listenkcp"`
	ListenSocks string `json:"listensocks"`
	FakeSubnet  string `json:"fakesubnet"`
	LocalSubnet string `json:"localsubnet"`
	AllowProxy  bool   `json:"allowproxy"`
	Key         string `json:"key"`
}

func parseConfig(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
