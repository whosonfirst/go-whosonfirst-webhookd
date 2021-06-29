package gocloud

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"gocloud.dev/blob"
	"net/url"
	"time"
)

func init() {

	ctx := context.Background()

	for _, scheme := range blob.DefaultURLMux().BucketSchemes() {

		err := dispatcher.RegisterDispatcher(ctx, scheme, NewBlobDispatcher)

		if err != nil {
			panic(err)
		}
	}

}

type BlobDispatcher struct {
	webhookd.WebhookDispatcher
	bucket *blob.Bucket
	prefix string
}

func NewBlobDispatcher(ctx context.Context, uri string) (webhookd.WebhookDispatcher, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()
	prefix := q.Get("dispatch_prefix")

	// START OF gocloud.dev/blob gets upset with flags it doesn't recognize

	q.Del("dispatch_prefix")
	u.RawQuery = q.Encode()
	uri = u.String()

	// END OF gocloud.dev/blob gets upset with flags it doesn't recognize

	bucket, err := blob.OpenBucket(ctx, uri)

	if err != nil {
		return nil, err
	}

	d := BlobDispatcher{
		bucket: bucket,
		prefix: prefix,
	}

	return &d, nil
}

func (d *BlobDispatcher) Dispatch(ctx context.Context, body []byte) *webhookd.WebhookError {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	fname, err := HashBody(ctx, body)

	if err != nil {
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	if d.prefix != "" {

		switch d.prefix {

		case "_ts_":

			now := time.Now()
			ts := now.Unix()

			fname = fmt.Sprintf("%d-%s", ts, fname)

		default:
			err := fmt.Errorf("Custom prefixes are not immplemented yet")
			return &webhookd.WebhookError{Code: 999, Message: err.Error()}
		}
	}

	wr, err := d.bucket.NewWriter(ctx, fname, nil)

	if err != nil {
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	_, err = wr.Write(body)

	if err != nil {
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	err = wr.Close()

	if err != nil {
		return &webhookd.WebhookError{Code: 999, Message: err.Error()}
	}

	return nil
}
