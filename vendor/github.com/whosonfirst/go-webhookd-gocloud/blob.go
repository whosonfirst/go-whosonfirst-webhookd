package gocloud

import (
	"context"
	"github.com/whosonfirst/go-webhookd/v3"
	"github.com/whosonfirst/go-webhookd/v3/dispatcher"
	"gocloud.dev/blob"
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
}

func NewBlobDispatcher(ctx context.Context, uri string) (webhookd.WebhookDispatcher, error) {

	bucket, err := blob.OpenBucket(ctx, uri)

	if err != nil {
		return nil, err
	}

	d := BlobDispatcher{
		bucket: bucket,
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
