package main

import (
	"log"
	"strings"

	"gopkg.in/go-playground/webhooks.v3"
	gh_webhooks "gopkg.in/go-playground/webhooks.v3/github"
)

func handleGitHubEvent(payload interface{}, header webhooks.Header) {
	log.Printf("DEBUG: Received %s event\n", header["X-Github-Event"])

	switch payload.(type) {
	case gh_webhooks.InstallationPayload:
		handleIntegrationInstallation(payload, header)
	case gh_webhooks.PullRequestPayload:
		handlePullRequest(payload, header)
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

		initInstallation()
	}

	saveConfig(&config)
}

// handlePullRequest handles GitHub pull_request events
func handlePullRequest(payload interface{}, header webhooks.Header) {
	pl := payload.(gh_webhooks.PullRequestPayload)

	switch pl.Action {
	case "opened":
		log.Printf("INFO: Handling Pull Request #%d\n", pl.PullRequest.Number)

		repoInfo, found := config.RepositoriesMap[repoID(pl.Repository.ID)]
		if !found {
			log.Printf("DEBUG: PR is not for a known repo (applies to %s), ignoring.\n", pl.Repository.Name)
			return
		}

		// If a new PR is open, assign it to the REVIEW column
		assignIssueToReview(ctx, repoInfo, pl.Number)
	}
}
