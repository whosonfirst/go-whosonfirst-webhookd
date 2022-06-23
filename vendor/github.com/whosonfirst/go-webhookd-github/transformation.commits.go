package github

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	gogithub "github.com/google/go-github/github"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/transformation"
	"net/url"
	"strconv"
)

func init() {

	ctx := context.Background()
	err := transformation.RegisterTransformation(ctx, "githubcommits", NewGitHubCommitsTransformation)

	if err != nil {
		panic(err)
	}
}

// see also: https://github.com/whosonfirst/go-whosonfirst-updated/issues/8

// GitHubCommitsTransformation implements the `webhookd.WebhookTransformation` interface for transforming GitHub
// commit webhook messages in to CSV data containing: the commit hash, the name of the repository and the path
// to the file commited.
type GitHubCommitsTransformation struct {
	webhookd.WebhookTransformation
	// ExcludeAdditions is a boolean flag to exclude newly added files from the final output.
	ExcludeAdditions bool
	// ExcludeModifications is a boolean flag to exclude updated (modified) files from the final output.
	ExcludeModifications bool
	// ExcludeDeletions is a boolean flag to exclude deleted files from the final output.
	ExcludeDeletions bool
}

// NewGitHubCommitsTransformation() creates a new `GitHubCommitsTransformation` instance, configured by 'uri'
// which is expected to take the form of:
//
//	githubcommits://?{PARAMETERS}
//
// Where {PARAMTERS} is:
// * `?exclude_additions` An optional boolean value to exclude newly added files from the final output.
// * `?exclude_modifications` An optional boolean value to exclude update (modified) files from the final output.
// * `?exclude_deletions` An optional boolean value to exclude deleted files from the final output.
func NewGitHubCommitsTransformation(ctx context.Context, uri string) (webhookd.WebhookTransformation, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	str_additions := q.Get("exclude_additions")
	str_modifications := q.Get("exclude_modifications")
	str_deletions := q.Get("exclude_deletions")

	exclude_additions := false
	exclude_modifications := false
	exclude_deletions := false

	if str_additions != "" {

		v, err := strconv.ParseBool(str_additions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_additions, err)
		}

		exclude_additions = v
	}

	if str_modifications != "" {

		v, err := strconv.ParseBool(str_modifications)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_modifications, err)
		}

		exclude_modifications = v
	}

	if str_deletions != "" {

		v, err := strconv.ParseBool(str_deletions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", str_deletions, err)
		}

		exclude_deletions = v
	}

	p := GitHubCommitsTransformation{
		ExcludeAdditions:     exclude_additions,
		ExcludeModifications: exclude_modifications,
		ExcludeDeletions:     exclude_deletions,
	}

	return &p, nil
}

// Transform() transforms 'body' (which is assumed to be a GitHub commit webhook message) in to CSV data containing:
// the commit hash, the name of the repository and the path to the file commited.
func (p *GitHubCommitsTransformation) Transform(ctx context.Context, body []byte) ([]byte, *webhookd.WebhookError) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	var event gogithub.PushEvent

	err := json.Unmarshal(body, &event)

	if err != nil {
		err := &webhookd.WebhookError{Code: 999, Message: err.Error()}
		return nil, err
	}

	buf := new(bytes.Buffer)
	wr := csv.NewWriter(buf)

	repo := event.Repo
	repo_name := *repo.Name
	commit_hash := *event.HeadCommit.ID

	for _, c := range event.Commits {

		if !p.ExcludeAdditions {
			for _, path := range c.Added {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}

		if !p.ExcludeModifications {
			for _, path := range c.Modified {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}

		if !p.ExcludeDeletions {
			for _, path := range c.Removed {
				commit := []string{commit_hash, repo_name, path}
				wr.Write(commit)
			}
		}
	}

	wr.Flush()

	return buf.Bytes(), nil
}
