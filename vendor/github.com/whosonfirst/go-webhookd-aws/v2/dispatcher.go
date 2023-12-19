package aws

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"log"
	"net/url"
	"regexp"
	"strings"
)

var preamble_re *regexp.Regexp

func init() {

	ctx := context.Background()
	err := dispatcher.RegisterDispatcher(ctx, "lambda", NewLambdaDispatcher)

	if err != nil {
		panic(err)
	}

	preamble_re = regexp.MustCompile(`^#\s?([^\s]+)\s(.*)$`)
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
	// An optional regular expression that will be compared to the commit message; if it matches the dispatcher will return an error with code `webhookd.HaltEvent`
	halt_on_message *regexp.Regexp
	// An optional regular expression that will be compared to the commit author; if it matches the dispatcher will return an error with code `webhookd.HaltEvent`
	halt_on_author *regexp.Regexp
}

// NewLambdaDispatcher returns a new `LambdaDispatcher` instance configured by 'uri' in the form of:
//
//	lambda://{FUNCTION_NAME}?{PARAMETERS}
//
// Where {PARAMETERS} are:
// * `dsn=` A valid `aaronland/go-aws-session` string used to create an AWS session instance.
// * `invocation_type=` The name of AWS Lambda invocation type. Valid options are: RequestResponse, Event, DryRun.
// * `?halt_on_message` An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent`
// * `?halt_on_author` An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent`
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

	q_halt_on_message := q.Get("halt_on_message")
	q_halt_on_author := q.Get("halt_on_author")

	if q_halt_on_message != "" {

		r, err := regexp.Compile(q_halt_on_message)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?halt_on_message= parameter, %w", err)
		}

		d.halt_on_message = r
	}

	if q_halt_on_author != "" {

		r, err := regexp.Compile(q_halt_on_author)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?halt_on_author= parameter, %w", err)
		}

		d.halt_on_author = r
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

	body, err := d.processBody(ctx, body)

	if err != nil {
		return err.(*webhookd.WebhookError)
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

func (d *LambdaDispatcher) processBody(ctx context.Context, body []byte) ([]byte, error) {

	if d.halt_on_message == nil && d.halt_on_author == nil {
		return body, nil
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	var message string
	var author string

	br := bytes.NewReader(body)
	scanner := bufio.NewScanner(br)

	for scanner.Scan() {

		ln := scanner.Text()
		m := preamble_re.FindStringSubmatch(ln)

		if len(m) != 3 {

			if strings.HasPrefix(ln, "#") {
				log.Printf("Unhandled comment '%s'", ln)
			}

			wr.WriteString(ln)
			continue
		}

		switch m[1] {
		case "message":

			message = m[2]

			if d.halt_on_message != nil && d.halt_on_message.MatchString(message) {
				return nil, &webhookd.WebhookError{Code: webhookd.HaltEvent, Message: "Halt"}
			}

		case "author":
			author = m[2]

			if d.halt_on_author != nil && d.halt_on_author.MatchString(author) {
				return nil, &webhookd.WebhookError{Code: webhookd.HaltEvent, Message: "Halt"}
			}

		default:
			log.Printf("Unhandled preamble '%s'", ln)
		}
	}

	wr.Flush()
	body = buf.Bytes()

	return body, nil
}
