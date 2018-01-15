package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"gopkg.in/go-playground/webhooks.v3"
	gh_webhooks "gopkg.in/go-playground/webhooks.v3/github"
)

const (
	path          = "/webhook"
	port          = 3016
	webhookSecret = "MyGitHubSuperSecretSecrect...?"
)

const (
	accessToken      = "e38f06927ca33e0fd7eab143b5d832536a2e88b3"
	repoOwner        = "PombeirP"
	repoName         = "status-github-bot"
	projectNumber    = 1
	reviewColumnName = "REVIEW"
)

var (
	client *github.Client
	ctx    context.Context
)

var projectID int

func main() {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)

	projects, _, err := client.Repositories.ListProjects(ctx, repoOwner, repoName, &github.ProjectListOptions{State: "open"})
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting")
		os.Exit(1)
	}
	for _, project := range projects {
		if project.GetNumber() == projectNumber {
			projectID = project.GetID()
		}
	}
	if projectID == 0 {
		fmt.Printf("FATAL: Could not find project #%d. Exiting\n", projectNumber)
		os.Exit(1)
	}

	fmt.Printf("Found project #%d. ID is %d\n", projectNumber, projectID)

	hook := gh_webhooks.New(&gh_webhooks.Config{Secret: webhookSecret})
	hook.RegisterEvents(HandlePullRequest, gh_webhooks.PullRequestEvent)

	err = webhooks.Run(hook, ":"+strconv.Itoa(port), path)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting")
		os.Exit(1)
	}
}

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header) {
	pl := payload.(gh_webhooks.PullRequestPayload)

	fmt.Printf("INFO: Handling Pull Request #%d\n", pl.PullRequest.Number)
	if pl.Action == "opened" {
		if pl.Repository.Owner.Login != repoOwner {
			fmt.Printf("DEBUG: PR is for repo owner %s, doesn't match watched repo owner %s\n", pl.Repository.Owner.Login, repoOwner)
			return
		}
		if pl.Repository.Name != repoName {
			fmt.Printf("DEBUG: PR is for repo %s, doesn't match watched repo %s\n", pl.Repository.Name, repoName)
			return
		}

		issue, _, err := client.Issues.Get(ctx, repoOwner, repoName, int(pl.Number))
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
		fmt.Printf("DEBUG: Fetched issue #%d. Will use ID %d", pl.Number, issue.GetID())

		columns, _, err := client.Projects.ListProjectColumns(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
		reviewColumn := getColumnFromName(columns, reviewColumnName)
		if reviewColumn == nil {
			fmt.Printf("ERROR: Could not find %s column in project board\n", reviewColumnName)
			return
		}

		fmt.Printf("INFO: Creating project card for PR %d\n", issue.GetID())
		// Create project card for the PR in the REVIEW column
		projectCardOptions := github.ProjectCardOptions{
			ContentID:   issue.GetID(),
			ContentType: "Issue",
		}
		card, _, err := client.Projects.CreateProjectCard(ctx, reviewColumn.GetID(), &projectCardOptions)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
		fmt.Printf("INFO: Created card %s", card.GetURL())
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
