package dispatcher

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	cache_blob "github.com/whosonfirst/go-cache-blob"
	webhookd "github.com/whosonfirst/go-webhookd/v3"
	webhookd_dispatcher "github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"gocloud.dev/blob"
	"io"
	"log"
	"net/url"
)

func init() {

	ctx := context.Background()
	err := webhookd_dispatcher.RegisterDispatcher(ctx, "findingaid-repo", NewFindingAidRepoDispatcher)

	if err != nil {
		panic(err)
	}
}

// FindingAidRepoDispatcher implements the webhookd.WebHookDispatcher and takes as input a []byte containing
// CSV-encoded rows produced by the go-webhookd-github.GitHubCommitsTransformation package and creates a
// corresponding go-whosonfirst-findingaid/repo.FindingAidResponse record for each row.
type FindingAidRepoDispatcher struct {
	webhookd.WebhookDispatcher
	indexer findingaid.Indexer
	acl     string
}

func NewFindingAidRepoDispatcher(ctx context.Context, uri string) (webhookd.WebhookDispatcher, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	cache_uri := q.Get("cache")

	if cache_uri == "" {
		return nil, fmt.Errorf("missing ?cache= parameter")
	}

	idx_q := url.Values{}

	idx_q.Set("cache", cache_uri)    // where the findingaid records will be written to
	idx_q.Set("iterator", "null://") // no-op

	idx_u := url.URL{}
	idx_u.Scheme = "repo"
	idx_u.RawQuery = idx_q.Encode()

	idx_uri := idx_u.String()

	indexer, err := findingaid.NewIndexer(ctx, idx_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create indexer for '%s', %w", idx_uri, err)
	}

	acl := q.Get("acl")

	d := FindingAidRepoDispatcher{
		indexer: indexer,
		acl:     acl,
	}

	return &d, nil
}

// Dispatch takes as input a []byte containing CSV-encoded rows produced by the go-webhookd-github.GitHubCommitsTransformation
// package and updates (or creates) a corresponding go-whosonfirst-findingaid record for each row.
func (d *FindingAidRepoDispatcher) Dispatch(ctx context.Context, body []byte) *webhookd.WebhookError {

	log.Println("DISPATCH", d.acl)
	
	// START OF S3 permissions stuff
	// This should really be moved in to a generic method like:
	// ctx = SetACLContextForS3Blob(ctx, "relevent-key-name", acl)
	
	if d.acl != "" {

		before := func(asFunc func(interface{}) bool) error {

			req := &s3manager.UploadInput{}
			ok := asFunc(&req)

			if !ok {
				return fmt.Errorf("invalid s3 type")
			}

			req.ACL = aws.String(d.acl)
			return nil
		}

		wr_opts := &blob.WriterOptions{
			BeforeWrite: before,
		}

		// This gets retrieved in whosonfirst/go-cache-blob.Set()
		ctx = context.WithValue(ctx, cache_blob.BlobCacheOptionsKey("options"), wr_opts)
	}

	// END OF S3 permissions stuff

	br := bytes.NewReader(body)
	csv_r := csv.NewReader(br)

	for {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

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

func (d *FindingAidRepoDispatcher) dispatchRow(ctx context.Context, row []string) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	count_cols := len(row)
	count_expected := 3

	if count_cols != count_expected {
		return fmt.Errorf("Invalid column count for row. Expected %d columns but row has %d", count_expected, count_cols)
	}

	// TBD: More validation (go-whosonfirst-repo?) or just assume that if
	// path is parseable, below, that everything else is probably okay?

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
