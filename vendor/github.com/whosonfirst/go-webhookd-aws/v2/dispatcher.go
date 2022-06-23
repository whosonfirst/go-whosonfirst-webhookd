package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"net/url"
)

func init() {

	ctx := context.Background()
	err := dispatcher.RegisterDispatcher(ctx, "lambda", NewLambdaDispatcher)

	if err != nil {
		panic(err)
	}
}

// LambdaDispatcher implements the `webhookd.WebhookDispatcher` interface for dispatching messages to an AWS Lambda function.
type LambdaDispatcher struct {
	webhookd.WebhookDispatcher
	// LambdaFunction is the name of the Lambda function to invoke.
	LambdaFunction string
	// LambdaService is `aws-sdk-go/service/lambda.Lambda` instance use to invoke a Lambda function.
	LambdaService *lambda.Lambda
	// invocation_type is the name of AWS Lambda invocation type.
	invocation_type string
}

// NewLambdaDispatcher returns a new `LambdaDispatcher` instance configured by 'uri' in the form of:
//
// 	lambda://{FUNCTION_NAME}?{PARAMETERS}
//
// Where {PARAMETERS} are:
// * `dsn=` A valid `aaronland/go-aws-session` string used to create an AWS session instance.
// * `invocation_type=` The name of AWS Lambda invocation type. Valid options are: RequestResponse, Event, DryRun.
func NewLambdaDispatcher(ctx context.Context, uri string) (webhookd.WebhookDispatcher, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	lambda_function := u.Host

	q := u.Query()

	lambda_dsn := q.Get("dsn")

	lambda_sess, err := session.NewSessionWithDSN(lambda_dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new AWS session, %w", err)
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

	d := LambdaDispatcher{
		LambdaFunction:  lambda_function,
		LambdaService:   lambda_svc,
		invocation_type: invocation_type,
	}

	return &d, nil
}

// Dispatch() relays 'body' as base64-endoded JSON string to the AWS Lambda function defined when 'd' was instantiated.
func (d *LambdaDispatcher) Dispatch(ctx context.Context, body []byte) *webhookd.WebhookError {

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
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(d.LambdaFunction),
		Payload:        payload,
		InvocationType: aws.String(d.invocation_type),
	}

	_, err = d.LambdaService.Invoke(input)

	if err != nil {
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	return nil
}
