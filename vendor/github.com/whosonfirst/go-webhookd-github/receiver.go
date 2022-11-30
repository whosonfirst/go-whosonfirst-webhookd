package github

// https://developer.github.com/webhooks/
// https://developer.github.com/webhooks/#payloads
// https://developer.github.com/v3/activity/events/types/#pushevent
// https://developer.github.com/v3/repos/hooks/#ping-a-hook

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	gogithub "github.com/google/go-github/v48/github"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/receiver"
	"io"
	_ "log"
	"net/http"
	"net/url"
)

func init() {

	ctx := context.Background()
	err := receiver.RegisterReceiver(ctx, "github", NewGitHubReceiver)

	if err != nil {
		panic(err)
	}
}

// GitHubReceiver implements the `webhookd.WebhookReceiver` interface for receiving webhook messages from GitHub.
type GitHubReceiver struct {
	webhookd.WebhookReceiver
	// secret is the shared secret used to generate signatures to validate messages.
	secret string
	// ref is the branch (reference) for which messages will be processed. Optional.
	ref string
}

// NewGitHubReceiver instantiates a new `GitHubReceiver` for receiving webhook messages from GitHub, configured
// by 'uri' which is expected to take the form of:
//
//	github://?secret={SECRET}&ref={BRANCH}
//
// Where {SECRET} is the shared secret used to generate signatures to validate messages and {BRANCH} is the optional
// branch (reference) name to limit message processing to.
func NewGitHubReceiver(ctx context.Context, uri string) (webhookd.WebhookReceiver, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	secret := q.Get("secret")
	ref := q.Get("ref")

	wh := GitHubReceiver{
		secret: secret,
		ref:    ref,
	}

	return wh, nil
}

// Receive() returns the body of the message in 'req'. It ensures that messages are sent as HTTP `POST` requests,
// that both `X-GitHub-Event` and `X-Hub-Signature` headers are present, that message body produces a valid signature
// using the secret used to create 'wh' and, if necessary, that the message is associated with the branch used to
// create 'wh'.
func (wh GitHubReceiver) Receive(ctx context.Context, req *http.Request) ([]byte, *webhookd.WebhookError) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	if req.Method != "POST" {

		code := http.StatusMethodNotAllowed
		message := "Method not allowed"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	event_type := req.Header.Get("X-GitHub-Event")

	if event_type == "" {

		code := http.StatusBadRequest
		message := "Bad Request - Missing X-GitHub-Event Header"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	sig := req.Header.Get("X-Hub-Signature")

	if sig == "" {

		code := http.StatusForbidden
		message := "Missing X-Hub-Signature required for HMAC verification"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	if event_type == "ping" {
		err := &webhookd.WebhookError{Code: webhookd.UnhandledEvent, Message: "ping message is a no-op"}
		return nil, err
	}

	// remember that you want to configure GitHub to send webhooks as 'application/json'
	// or all this code will get confused (20190212/thisisaaronland)

	body, err := io.ReadAll(req.Body)

	if err != nil {

		code := http.StatusInternalServerError
		message := err.Error()

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	expectedSig, _ := GenerateSignature(string(body), wh.secret)

	if !hmac.Equal([]byte(expectedSig), []byte(sig)) {

		code := http.StatusForbidden
		message := "HMAC verification failed"

		err := &webhookd.WebhookError{Code: code, Message: message}
		return nil, err
	}

	if wh.ref != "" {

		var event gogithub.PushEvent

		err := json.Unmarshal(body, &event)

		if err != nil {
			err := &webhookd.WebhookError{Code: 999, Message: err.Error()}
			return nil, err
		}

		if wh.ref != *event.Ref {

			msg := "Invalid ref for commit"
			err := &webhookd.WebhookError{Code: 666, Message: msg}
			return nil, err
		}
	}

	/*

		So here's a thing that's not awesome: the event_type is passed in the header
		rather than anywhere in the payload body. So I don't know... maybe we need to
		change the signature of Receive method to be something like this:
		       { Payload: []byte, Extras: map[string]string }

		Which is not something that makes me "happy"... (20161016/thisisaaronland)

	*/

	return body, nil
}
