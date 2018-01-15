package main

import (
	"log"
	"strings"

	"gopkg.in/go-playground/webhooks.v3"
	gh_webhooks "gopkg.in/go-playground/webhooks.v3/github"
)

func handleGitHubEvent(payload interface{}, header webhooks.Header) {
	log.Printf("DEBUG: Received %s event\n", header["X-Github-Event"])

	switch pl := payload.(type) {
	case gh_webhooks.InstallationPayload:
		handleIntegrationInstallation(payload, header)
	case gh_webhooks.PullRequestPayload: // TODO: Add types required for new use cases in this case statement
		repoID := repoID(pl.Repository.ID)
		repoName := pl.Repository.Name

		repoInfo, found := config.RepositoriesMap[repoID]
		if !found {
			log.Printf("DEBUG: Event is not for a known repo (applies to %s), ignoring.\n", repoName)
			return
		}

		// Allow any usecases to execute on this event
		for _, useCase := range useCases {
			useCase.Execute(ctx, repoInfo, payload)
		}
	}
}

// handleInstallation handles GitHub ping events
func handleIntegrationInstallation(payload interface{}, header webhooks.Header) {
	installationPayload := payload.(gh_webhooks.InstallationPayload)

	log.Printf("INFO: Handling Integration Installation %s event for app with ID #%d\n", installationPayload.Action, installationPayload.Installation.AppID)

	switch installationPayload.Action {
	case "deleted":
		config.InstallationID = 0
		config.ApplicationID = 0
	case "created":
		config.InstallationID = int(installationPayload.Installation.ID)
		config.ApplicationID = installationPayload.Installation.AppID

		if installationPayload.Installation.Permissions.PullRequests != "read" {
			log.Println("WARN: The app is missing pull_request permission. Please set it to READ.")
		}

		if len(installationPayload.Repositories) == 0 {
			log.Fatal("FATAL: The app is not configured on any repositories. Please reinstall the app with some repositories.")
		}

		isSubscribedToPREvents := false
		for _, evt := range installationPayload.Installation.Events {
			if evt == "pull_request" {
				isSubscribedToPREvents = true
				break
			}
		}
		if !isSubscribedToPREvents {
			log.Fatal("WARN: The app needs to be subscribed to pull_request events. Please reinstall app with it.")
		}

		// Populate RepositoriesMap before calling initInstallation
		for _, repo := range installationPayload.Repositories {
			config.RepositoriesMap[repoID(repo.ID)] = &repoInfo{ID: repoID(repo.ID), Owner: strings.Split(repo.FullName, "/")[0], Name: repo.Name}
		}

		if err := initInstallation(ctx); err != nil {
			return
		}
	}

	saveConfig(&config)
}
