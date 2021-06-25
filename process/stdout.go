package process

import (
	"context"
	"fmt"
	"os"
)

type StdoutProcessor struct {
	Processor
}

func init() {
	ctx := context.Background()
	RegisterProcessor(ctx, "stdout", NewStdoutProcessor)
}

func NewStdoutProcessor(ctx context.Context, uri string) (Processor, error) {

	pr := &StdoutProcessor{}
	return pr, nil
}

func (pr *StdoutProcessor) Process(ctx context.Context, repo string, files ...string) error {

	for _, f := range files {
		fmt.Fprintf(os.Stdout, "%s\t%s\n", repo, f)
	}

	return nil
}
