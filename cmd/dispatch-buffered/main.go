// Crawl an S3 directory of files containing WOF repos to index.
// These are assumed to have been written using the go-webhookd(-gocloud) package.
// Specifically the repo name will have been extracted using the https://github.com/whosonfirst/go-webhookd-github#github
// receiver and the https://github.com/whosonfirst/go-webhookd-github#githubrepo transformation
// and finally written to S3 using the https://github.com/whosonfirst/go-webhookd-gocloud dispatcher.
// This tool will invoke a Lambda function, which is not necessarily but otherwise assumed to be the one
// defined in cmd/launch-ecs-task which, in turn, will invoke an ECS task. The point being it's just a giant
// bucket brigade to pass around a repo name in order to prevent invoking a thundering herd of ECS tasks when a repo
// has a lot of small atomic commits (triggering GitHub receiver above).
//
// For example:
// $> ./bin/dispatch-buffered -dryrun \
//	-bucket-uri 's3blob://sfomuseum-pending/?region=us-west-2&prefix=indexing/&credentials=session' \
//	-lambda-uri 'lambda://WebhookdIndexingTask?dsn=region=us-west-2 credentials=session'
package main

import (
	_ "github.com/aaronland/gocloud-blob-s3"
)

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aaronland/go-aws-session"
	go_lambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"gocloud.dev/blob"
	"io"
	"log"
	"net/url"
	"strings"
)

// START OF please merge with aaronland/go-aws-lambda

type Dispatcher struct {
	LambdaFunction  string
	LambdaService   *lambda.Lambda
	invocation_type string
}

func (d *Dispatcher) Dispatch(ctx context.Context, body []byte) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	// I don't understand why I need to base64 encode this...
	// (20200526/thisisaaronland)

	enc_body := base64.StdEncoding.EncodeToString(body)

	payload, err := json.Marshal(enc_body)

	if err != nil {
		return fmt.Errorf("Failed to marshal body, %w", err)
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(d.LambdaFunction),
		Payload:        payload,
		InvocationType: aws.String(d.invocation_type),
	}

	_, err = d.LambdaService.Invoke(input)

	if err != nil {
		return fmt.Errorf("Failed to invoke service (%s), %w", d.LambdaFunction, err)
	}

	return nil
}

func NewDispatcher(ctx context.Context, uri string) (*Dispatcher, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse Lambda URI, %s", err)
	}

	lambda_function := u.Host

	q := u.Query()

	lambda_dsn := q.Get("dsn")

	lambda_sess, err := session.NewSessionWithDSN(lambda_dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new session, %v", err)
	}

	invocation_type := q.Get("invocation_type")

	switch invocation_type {
	case "":
		invocation_type = "RequestResponse"
	case "RequestResponse", "Event", "DryRun":
		// pass
	default:
		return nil, fmt.Errorf("Invalid invocation_type parameter")
	}

	lambda_svc := lambda.New(lambda_sess)

	d := &Dispatcher{
		LambdaFunction:  lambda_function,
		LambdaService:   lambda_svc,
		invocation_type: invocation_type,
	}

	return d, nil
}

// END OF please merge with aaronland/go-aws-lambda

func main() {

	fs := flagset.NewFlagSet("dispatch")

	var lambda_uris multi.MultiString
	fs.Var(&lambda_uris, "lambda-uri", "One or more valid aaronland/go-aws-lambda URIs. If run in -mode lambda mulitple values can be specified as a ';' separater list in the `WEBHOOKD_LAMBDA_URIS` environment variable.")

	bucket_uri := fs.String("bucket-uri", "", "A valid gocloud.dev/blob Bucket URI where buffered dispatch messages are stored.")

	mode := fs.String("mode", "cli", "Valid options are: cli, lambda.")
	dryrun := fs.Bool("dryrun", false, "Go through the motions but don't invoke any tasks")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "WEBHOOKD", true)

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	var dispatchers []*Dispatcher

	if *mode == "lambda" {
		lambda_uris = strings.Split(lambda_uris[0], ";")
	}

	for _, uri := range lambda_uris {

		d, err := NewDispatcher(ctx, uri)

		if err != nil {
			log.Fatalf("Failed to create dispatcher for '%s', %v", uri, err)
		}

		dispatchers = append(dispatchers, d)
	}

	switch *mode {
	case "cli":

		err = process(ctx, bucket, *dryrun, dispatchers...)

		if err != nil {
			log.Fatalf("Failed to process bucket, %v", err)
		}

	case "lambda":

		handler := func(ctx context.Context) error {
			return process(ctx, bucket, *dryrun, dispatchers...)
		}

		go_lambda.Start(handler)

	default:

		log.Fatalf("Invalid mode")
	}
}

func process(ctx context.Context, bucket *blob.Bucket, dryrun bool, dispatchers ...*Dispatcher) error {

	var list func(context.Context, *blob.Bucket, string) error

	list = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		iter := bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {
			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("Failed to iterate bucket, %v", err)
			}

			path := obj.Key

			if obj.IsDir {

				err := list(ctx, bucket, path)

				if err != nil {
					return err
				}

				continue
			}

			fh, err := bucket.NewReader(ctx, path, nil)

			if err != nil {
				return fmt.Errorf("Failed to open '%s', %v", path, err)
			}

			defer fh.Close()

			body, err := io.ReadAll(fh)

			if err != nil {
				return fmt.Errorf("Failed to read '%s', %v", path, err)
			}

			if dryrun {

				for _, d := range dispatchers {
					log.Printf("Dispatch '%s' to %v\n", string(body), d)
				}

				return nil
			}

			for _, d := range dispatchers {

				err := d.Dispatch(ctx, body)

				if err != nil {
					return fmt.Errorf("Failed to dispatch function, %v", err)
				}
			}

			fh.Close()

			err = bucket.Delete(ctx, path)

			if err != nil {
				return fmt.Errorf("Failed to remove %s, %v", path, err)
			}
		}

		return nil
	}

	err := list(ctx, bucket, "")

	if err != nil {
		log.Fatalf("Failed to list bucket, %v", err)
	}

	return nil
}
