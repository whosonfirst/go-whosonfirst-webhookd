package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gogithub "github.com/google/go-github/v48/github"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/transformation"
	"net/url"
	"strconv"
)

func init() {

	ctx := context.Background()
	err := transformation.RegisterTransformation(ctx, "githubrepo", NewGitHubRepoTransformation)

	if err != nil {
		panic(err)
	}
}

// see also: https://github.com/whosonfirst/go-whosonfirst-updated/issues/8

// GitHubRepoTransformation implements the `webhookd.WebhookTransformation` interface for transforming GitHub
// commit webhook messages in to the name of the repository where the commit occurred.
type GitHubRepoTransformation struct {
	webhookd.WebhookTransformation
	// ExcludeAdditions is a boolean flag to exclude newly added files from consideration.
	ExcludeAdditions bool
	// ExcludeModifications is a boolean flag to exclude updated (modified) files from consideration.
	ExcludeModifications bool
	// ExcludeDeletions is a boolean flag to exclude updated (modified) files from consideration.
	ExcludeDeletions bool
}

// NewGitHubRepoTransformation() creates a new `GitHubRepoTransformation` instance, configured by 'uri'
// which is expected to take the form of:
//
//	githubrepo://?{PARAMETERS}
//
// Where {PARAMTERS} is:
// * `?exclude_additions` An optional boolean value to exclude newly added files from consideration.
// * `?exclude_modifications` An optional boolean value to exclude update (modified) files from consideration.
// * `?exclude_deletions` An optional boolean value to exclude deleted files from consideration.
func NewGitHubRepoTransformation(ctx context.Context, uri string) (webhookd.WebhookTransformation, error) {

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

	p := GitHubRepoTransformation{
		ExcludeAdditions:     exclude_additions,
		ExcludeModifications: exclude_modifications,
		ExcludeDeletions:     exclude_deletions,
	}

	return &p, nil
}

// Transform() transforms 'body' (which is assumed to be a GitHub commit webhook message) in to name of the repository
// where the commit occurred.
func (p *GitHubRepoTransformation) Transform(ctx context.Context, body []byte) ([]byte, *webhookd.WebhookError) {

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

	repo := event.Repo
	repo_name := *repo.Name

	has_updates := false

	for _, c := range event.Commits {

		if !p.ExcludeAdditions {

			if len(c.Added) > 0 {
				has_updates = true
			}
		}

		if !p.ExcludeModifications {

			if len(c.Modified) > 0 {
				has_updates = true
			}
		}

		if !p.ExcludeDeletions {

			if len(c.Removed) > 0 {
				has_updates = true
			}
		}
	}

	if has_updates {
		buf.WriteString(repo_name)
	}

	return buf.Bytes(), nil
}
