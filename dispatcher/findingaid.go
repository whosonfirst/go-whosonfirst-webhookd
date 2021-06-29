package dispatcher

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	webhookd "github.com/whosonfirst/go-webhookd/v3"
	webhookd_dispatcher "github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"net/url"
)

func init() {

	ctx := context.Background()
	err := webhookd_dispatcher.RegisterDispatcher(ctx, "findingaid", NewFindingAidDispatcher)

	if err != nil {
		panic(err)
	}
}

type FindingAidDispatcher struct {
	webhookd.WebhookDispatcher
	indexer findingaid.Indexer
}

func NewFindingAidDispatcher(ctx context.Context, uri string) (webhookd.WebhookDispatcher, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	indexer_uri := q.Get("indexer")

	indexer, err := findingaid.NewIndexer(ctx, indexer_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create indexer for '%s', %w", indexer_uri, err)
	}

	d := FindingAidDispatcher{
		indexer: indexer,
	}

	return &d, nil
}

// Dispatch takes as input a []byte containing CSV-encoded rows produced by the go-webhookd-github.GitHubCommitsTransformation
// package and updates (or creates) a corresponding go-whosonfirst-findingaid record for each row.
func (d *FindingAidDispatcher) Dispatch(ctx context.Context, body []byte) *webhookd.WebhookError {

	br := bytes.NewReader(body)
	csv_r := csv.NewReader(br)

	for {
		row, err := csv_r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return &webhookd.WebhookError{Code: 999, Message: err.Error()}
		}

		err = d.dispatchRow(ctx, row)

		if err != nil {
			return &webhookd.WebhookError{Code: 999, Message: err.Error()}
		}

	}

	return nil
}

func (d *FindingAidDispatcher) dispatchRow(ctx context.Context, row []string) error {

	count_cols := len(row)
	count_expected := 3

	if count_cols != count_expected {
		return fmt.Errorf("Invalid column count for row. Expected %d columns but row has %d", count_expected, count_cols)
	}

	// PLEASE VALIDATE THESE...

	// commit_hash := row[0]
	repo_name := row[1]
	path := row[2]

	id, _, err := uri.ParseURI(path)

	if err != nil {
		return fmt.Errorf("Failed to parse '%s', %v", path, err)
	}

	// Basically we spoofing something that can be read by
	// go-whosonfirst-findingaid/repo.FindingAidResponseFromReader

	type Feature struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
	}

	props := map[string]interface{}{
		"wof:id":   id,
		"wof:repo": repo_name,
	}

	f := Feature{
		Type:       "Feature",
		Properties: props,
	}

	enc_f, err := json.Marshal(f)

	if err != nil {
		return fmt.Errorf("Failed to encode feature, %v", err)
	}

	fr := bytes.NewReader(enc_f)

	return d.indexer.IndexReader(ctx, fr)
}
