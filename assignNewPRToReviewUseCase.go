package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	gh_webhooks "gopkg.in/go-playground/webhooks.v3/github"
)

type assignNewPRToReviewUseCase struct {
}

func (u *assignNewPRToReviewUseCase) Init(ctx context.Context, repoInfo *repoInfo) error {
	if repoInfo.ReviewColumnID != 0 {
		// Already initialized from config.json
		return nil
	}

	columns, _, err := client.Projects.ListProjectColumns(ctx, repoInfo.ProjectID, nil)
	if err != nil {
		log.Printf("WARN: %s\n", err)
		return err
	}
	reviewColumn := getColumnFromName(columns, config.ReviewColumnName)
	if reviewColumn == nil {
		log.Printf("WARN: Could not find '%s' column in project board\n", config.ReviewColumnName)
		return err
	}
	repoInfo.ReviewColumnID = reviewColumn.GetID()
	log.Printf("DEBUG: Found project column '%s'. ID is %d\n", reviewColumn.GetName(), repoInfo.ReviewColumnID)

	return nil
}

func (u *assignNewPRToReviewUseCase) Execute(ctx context.Context, repoInfo *repoInfo, payload interface{}) error {
	switch pl := payload.(type) {
	case gh_webhooks.PullRequestPayload:
		switch pl.Action {
		case "opened":
			log.Printf("INFO: Handling Pull Request #%d\n", pl.PullRequest.Number)

			return assignIssueToReview(ctx, repoInfo, pl.PullRequest.Number)
		}
	}

	return nil
}

func assignIssueToReview(ctx context.Context, repoInfo *repoInfo, prNumber int64) error {
	issue, _, err := client.Issues.Get(ctx, repoInfo.Owner, repoInfo.Name, int(prNumber))
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return err
	}
	log.Printf("DEBUG: Fetched issue %d for PR %d", issue.GetID(), prNumber)

	// Create project card for the PR in the REVIEW column
	log.Printf("INFO: Creating project card for PR %d\n", issue.GetID())
	projectCardOptions := github.ProjectCardOptions{
		ContentID:   issue.GetID(),
		ContentType: "Issue",
	}
	card, _, err := client.Projects.CreateProjectCard(ctx, repoInfo.ReviewColumnID, &projectCardOptions)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return err
	}
	log.Printf("DEBUG: Created card %s", card.GetURL())

	return nil
}

func getColumnFromName(columns []*github.ProjectColumn, name string) *github.ProjectColumn {
	for _, column := range columns {
		if column.GetName() == name {
			return column
		}
	}

	return nil
}
