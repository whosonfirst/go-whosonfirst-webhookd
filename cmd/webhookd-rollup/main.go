// Rollup files produced by the go-webhookd-github 'githubcommits' transformer and written a to gocloud.dev/blob (go-webhookd-gocloud) dispatcher source.
package main

// It's possible that some or all of this code should be moved in to
// whosonfirst/go-webhookd-gocloud but it's still too soon for that

import (
	_ "github.com/aaronland/go-cloud-s3blob"
	_ "gocloud.dev/blob/fileblob"
)

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-webhookd/process"
	"github.com/whosonfirst/go-whosonfirst-webhookd/rollup"
	"gocloud.dev/blob"
	"log"
)

func RollupAndProcess(ctx context.Context, bucket *blob.Bucket, pr process.Processor) error {

	c, err := rollup.RollupBucket(ctx, bucket)

	if err != nil {
		return err
	}

	for repo, files := range c.Commits() {

		err := pr.Process(ctx, repo, files...)

		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	fs := flagset.NewFlagSet("webhookd")

	mode := fs.String("mode", "cli", "Valid options are: cli, lambda.")

	bucket_uri := fs.String("bucket-uri", "", "A valid gocloud.dev/blob (bucket) URI.")

	processor_uri := fs.String("processor-uri", "stdout://", "...")

	flagset.Parse(fs)

	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	pr, err := process.NewProcessor(ctx, *processor_uri)

	if err != nil {
		log.Fatalf("Failed to create processor, %v", err)
	}

	switch *mode {
	case "cli":

		err := RollupAndProcess(ctx, bucket, pr)

		if err != nil {
			log.Fatalf("Failed to rollup and dispatch, %v", err)
		}

	case "lambda":

		handler := func(ctx context.Context) error {
			return RollupAndProcess(ctx, bucket, pr)
		}

		lambda.Start(handler)

	default:
		log.Fatalf("Unsuported mode, '%s'", *mode)
	}

}
