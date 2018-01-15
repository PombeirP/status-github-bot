package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type repoID int64
type repoInfo struct {
	ID             repoID
	Owner          string
	Name           string
	ProjectID      int
	ReviewColumnID int
}

type installationConfig struct {
	Port             int                  `json:"port"`
	WebhookSecret    string               `json:"webhook_secret"`
	ProjectBoardName string               `json:"project_board_name"`
	ReviewColumnName string               `json:"review_column_name"`
	RepositoriesMap  map[repoID]*repoInfo `json:"repositories"`
	ApplicationID    int                  `json:"application_id"`
	InstallationID   int                  `json:"installation_id"`
}

func loadConfig() installationConfig {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("FATAL: %s. Exiting\n", err)
	}

	var c installationConfig
	json.Unmarshal(raw, &c)
	if c.RepositoriesMap == nil {
		c.RepositoriesMap = make(map[repoID]*repoInfo)
	}

	return c
}

func saveConfig(config *installationConfig) error {
	b, err := json.Marshal(config)
	if err != nil {
		log.Printf("WARN: Failed to marshal config object: %s\n", err)
		return err
	}

	err = ioutil.WriteFile("./config.json", b, os.ModeExclusive)
	if err != nil {
		log.Printf("WARN: Failed to save config file: %s\n", err)
	}

	return err
}
