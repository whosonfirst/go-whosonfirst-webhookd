package process

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"net/url"
	"sort"
	"strings"
)

type Processor interface {
	Process(context.Context, string, ...string) error
}

type ProcessorInitializeFunc func(ctx context.Context, uri string) (Processor, error)

var processors roster.Roster

func ensureSpatialRoster() error {

	if processors == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		processors = r
	}

	return nil
}

func RegisterProcessor(ctx context.Context, scheme string, f ProcessorInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return err
	}

	return processors.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range processors.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewProcessor(ctx context.Context, uri string) (Processor, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := processors.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(ProcessorInitializeFunc)
	return f(ctx, uri)
}
