package main

import (
	"encoding/json"
	"os"
)

type clientConfig struct {
	UsingGdrive  bool
	UsingDropbox bool
	UsingFlickr  bool
}

func loadClientConfig() *clientConfig {
	cfg := &clientConfig{}
	f, err := os.Open("horcrux_client_config.json")
	if err == nil {
		err = json.NewDecoder(f).Decode(cfg)
		defer f.Close()
	}
	logInfo.Printf("Loaded %v\n", cfg)
	return cfg
}

func (cfg *clientConfig) Save() error {
	file := "horcrux_client_config.json"
	logInfo.Printf("Saving client configuration to: %s\n", file)
	logInfo.Printf("Saving %v\n", cfg)
	f, err := os.Create(file)
	if err != nil {
		logError.Printf("Unable to cache oauth token: %v", err)
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(cfg)
	return nil
}

func (cfg *clientConfig) SetUsingGdrive() {
	cfg.UsingGdrive = true
	_ = cfg.Save()
}
func (cfg *clientConfig) SetUsingDropbox() {
	cfg.UsingDropbox = true
	_ = cfg.Save()
}

func (cfg *clientConfig) SetUsingFlickr() {
	cfg.UsingFlickr = true
	_ = cfg.Save()
}
