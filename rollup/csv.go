package rollup

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

func ProcessCSVFile(ctx context.Context, c *Catalog, fh io.Reader) error {

	/*

		To do: Read fh in to bytes and hash the body and check to see whether
		we've processed this file already. If not create a BytesReader and pass
		that to the CSV reader.

	*/

	csv_r := csv.NewReader(fh)

	for {
		row, err := csv_r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		count_cols := len(row)
		count_expected := 3

		if count_cols != count_expected {
			return fmt.Errorf("Invalid githubcommit CSV columns, expected %d but got %d", count_expected, count_cols)
		}

		repo := row[1]
		file := row[2]

		err = c.Add(ctx, repo, file)

		if err != nil {
			return err
		}

	}

	return nil
}
