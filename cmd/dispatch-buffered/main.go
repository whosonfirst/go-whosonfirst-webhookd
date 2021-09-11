// Crawl an S3 directory of files containing WOF repos to index.
// These are assumed to have been written using the go-webhookd(-gocloud) package.
// Specifically the repo name will have been extracted using the https://github.com/whosonfirst/go-webhookd-github#github
// receiver and the https://github.com/whosonfirst/go-webhookd-github#githubrepo transformation
// and finally written to S3 using the https://github.com/whosonfirst/go-webhookd-gocloud dispatcher.
// This tool will invoke a Lambda function, which is not necessarily but otherwise assumed to be the one
// defined in https://github.com/whosonfirst/go-webhookd-aws/tree/main/cmd/webhookd-aws-launch-task
// which, in turn, will invoke an ECS task. The point being it's just a giant bucket brigade to
// pass around a repo name in order to prevent invoking a thundering herd of ECS tasks when a repo
// has a lot of small atomic commits (triggering GitHub receiver above).
package main

import ()

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aaronland/go-aws-session"
	go_lambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/sfomuseum/go-flags/multi"
	"gocloud.dev/blob"
	"io"
	"log"
	"net/url"
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

	var lambda_uris multi.MultiString
	flag.Var(&lambda_uris, "lambda-uri", "...")

	bucket_uri := flag.String("bucket-uri", "", "...")

	mode := flag.String("mode", "cli", "...")

	flag.Parse()

	ctx := context.Background()

	bucket, err := blob.OpenBucket(ctx, *bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	var dispatchers []*Dispatcher

	for _, uri := range lambda_uris {

		d, err := NewDispatcher(ctx, uri)

		if err != nil {
			log.Fatalf("Failed to create dispatcher for '%s', %v", uri, err)
		}

		dispatchers = append(dispatchers, d)
	}

	switch *mode {
	case "cli":

		err = process(ctx, bucket, dispatchers...)

		if err != nil {
			log.Fatalf("Failed to process bucket, %v", err)
		}

	case "lambda":

		handler := func(ctx context.Context) error {
			return process(ctx, bucket, dispatchers...)
		}

		go_lambda.Start(handler)

	default:

		log.Fatalf("Invalid mode")
	}
}

func process(ctx context.Context, bucket *blob.Bucket, dispatchers ...*Dispatcher) error {

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
