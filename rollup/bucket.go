package rollup

import (
	"context"
	"gocloud.dev/blob"
	"io"
)

// Rollup takes all the files in 'bucket'
func RollupBucket(ctx context.Context, bucket *blob.Bucket) (*Catalog, error) {

	c, err := NewCatalog()

	if err != nil {
		return nil, err
	}

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
				return err
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
				return err
			}

			defer fh.Close()

			err = ProcessCSVFile(ctx, c, fh)

			if err != nil {
				return err
			}

		}

		return nil
	}

	err = list(ctx, bucket, "")

	if err != nil {
		return nil, err
	}

	return c, nil
}
