package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
)

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
