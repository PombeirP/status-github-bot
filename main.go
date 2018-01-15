package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"

	"gopkg.in/go-playground/webhooks.v3"
	gh_webhooks "gopkg.in/go-playground/webhooks.v3/github"
)

const (
	path = "/webhook"
)

var (
	client *github.Client
	ctx    context.Context
	config installationConfig
)

func main() {
	logger := &lumberjack.Logger{
		Filename:   "./bot.log",
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logger))

	ctx = context.Background()

	config = loadConfig()
	if config.InstallationID != 0 {
		initInstallation()
	}

	hook := gh_webhooks.New(&gh_webhooks.Config{Secret: config.WebhookSecret})
	hook.RegisterEvents(handleGitHubEvent, gh_webhooks.IntegrationInstallationEvent, gh_webhooks.PullRequestEvent)

	err := webhooks.Run(hook, ":"+strconv.Itoa(config.Port), path)
	if err != nil {
		log.Fatalf("FATAL: %s. Exiting\n", err)
	} else {
		saveConfig(&config)
	}
}

func initInstallation() {
	log.Printf("INFO: Initializing installation for app with ID #%d and installation ID #%d", config.ApplicationID, config.InstallationID)

	// Wrap the shared transport for use with the integration ID authenticating with installation ID.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, config.ApplicationID, config.InstallationID, "./status-github-bot.private-key.pem")
	if err != nil {
		log.Fatalf("FATAL: %s. Exiting\n", err)
	}

	// Use installation transport with client.
	client = github.NewClient(&http.Client{Transport: itr})

	for _, repoInfo := range config.RepositoriesMap {
		projects, _, err := client.Repositories.ListProjects(ctx, repoInfo.Owner, repoInfo.Name, &github.ProjectListOptions{State: "open"})
		if err != nil {
			log.Fatalf("FATAL: %s. Exiting\n", err)
		}
		var project *github.Project
		for _, project = range projects {
			if project.GetName() == config.ProjectBoardName {
				repoInfo.ProjectID = project.GetID()
			}
		}
		if repoInfo.ProjectID == 0 {
			log.Fatalf("FATAL: Could not find project named '%s'. Exiting\n", config.ProjectBoardName)
		}

		log.Printf("DEBUG: Found project named '%s'. ID is %d\n", project.GetName(), repoInfo.ProjectID)

		columns, _, err := client.Projects.ListProjectColumns(ctx, repoInfo.ProjectID, nil)
		if err != nil {
			log.Printf("ERROR: %s\n", err)
			return
		}
		reviewColumn := getColumnFromName(columns, config.ReviewColumnName)
		if reviewColumn == nil {
			log.Printf("ERROR: Could not find '%s' column in project board\n", config.ReviewColumnName)
			return
		}
		repoInfo.ReviewColumnID = reviewColumn.GetID()
		log.Printf("DEBUG: Found project column '%s'. ID is %d\n", reviewColumn.GetName(), repoInfo.ReviewColumnID)
	}
}

func getColumnFromName(columns []*github.ProjectColumn, name string) *github.ProjectColumn {
	for _, column := range columns {
		if column.GetName() == name {
			return column
		}
	}

	return nil
}
