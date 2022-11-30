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
	"regexp"
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
	// A boolean flag signaling the commit message should be prepended to the top of the final output in the form of '#message {COMMIT_MESSAGE}'
	prepend_message bool
	// A boolean flag signaling the commit author should be prepended to the top of the final output in the form of '#author {COMMIT_AUTHOR}'
	prepend_author bool
	// An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent`
	halt_on_message *regexp.Regexp
	// An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent`
	halt_on_author *regexp.Regexp
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
// * `?prepend_message` An optional boolean value to prepend the commit message to the final output. This takes the form of '#message {COMMIT_MESSAGE}'
// * `?prepend_author` An optional boolean value to prepend the name of the commit author to the final output. This takes the form of '#author {COMMIT_AUTHOR}'
// * `?halt_on_message` An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent`
// * `?halt_on_author` An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent`
func NewGitHubRepoTransformation(ctx context.Context, uri string) (webhookd.WebhookTransformation, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	q_additions := q.Get("exclude_additions")
	q_modifications := q.Get("exclude_modifications")
	q_deletions := q.Get("exclude_deletions")
	q_message := q.Get("prepend_message")
	q_author := q.Get("prepend_author")

	q_halt_on_message := q.Get("halt_on_message")
	q_halt_on_author := q.Get("halt_on_author")

	exclude_additions := false
	exclude_modifications := false
	exclude_deletions := false

	prepend_message := false
	prepend_author := false

	if q_additions != "" {

		v, err := strconv.ParseBool(q_additions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", q_additions, err)
		}

		exclude_additions = v
	}

	if q_modifications != "" {

		v, err := strconv.ParseBool(q_modifications)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", q_modifications, err)
		}

		exclude_modifications = v
	}

	if q_deletions != "" {

		v, err := strconv.ParseBool(q_deletions)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", q_deletions, err)
		}

		exclude_deletions = v
	}

	if q_message != "" {

		v, err := strconv.ParseBool(q_message)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", q_message, err)
		}

		prepend_message = v
	}

	if q_author != "" {

		v, err := strconv.ParseBool(q_author)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse '%s', %w", q_author, err)
		}

		prepend_author = v
	}

	p := GitHubRepoTransformation{
		ExcludeAdditions:     exclude_additions,
		ExcludeModifications: exclude_modifications,
		ExcludeDeletions:     exclude_deletions,
		prepend_message:      prepend_message,
		prepend_author:       prepend_author,
	}

	if q_halt_on_message != "" {

		r, err := regexp.Compile(q_halt_on_message)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?halt_on_message= parameter, %w", err)
		}

		p.halt_on_message = r
	}

	if q_halt_on_author != "" {

		r, err := regexp.Compile(q_halt_on_author)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?halt_on_author= parameter, %w", err)
		}

		p.halt_on_author = r
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

	if p.halt_on_message != nil && p.halt_on_message.MatchString(*event.HeadCommit.Message) {
		err := &webhookd.WebhookError{Code: webhookd.HaltEvent, Message: "Halt"}
		return nil, err
	}

	if p.halt_on_author != nil && p.halt_on_author.MatchString(*event.HeadCommit.Author.Name) {
		err := &webhookd.WebhookError{Code: webhookd.HaltEvent, Message: "Halt"}
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

		if p.prepend_message {
			msg := fmt.Sprintf("#message %s\n", *event.HeadCommit.Message)
			buf.WriteString(msg)
		}

		if p.prepend_author {
			msg := fmt.Sprintf("#author %s\n", *event.HeadCommit.Author.Name)
			buf.WriteString(msg)
		}

		buf.WriteString(repo_name)
	}

	return buf.Bytes(), nil
}
