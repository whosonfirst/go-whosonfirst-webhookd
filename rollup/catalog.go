package rollup

import (
	"context"
	"sync"
)

type Catalog struct {
	commits map[string][]string
	mu      *sync.RWMutex
}

func NewCatalog() (*Catalog, error) {

	mu := new(sync.RWMutex)
	commits := make(map[string][]string)

	c := &Catalog{
		mu:      mu,
		commits: commits,
	}

	return c, nil
}

func (c *Catalog) Add(ctx context.Context, repo string, file string) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	files, ok := c.commits[repo]

	if !ok {
		files = make([]string, 0)
	}

	files = append(files, file)
	c.commits[repo] = files

	return nil
}

func (c *Catalog) Commits() map[string][]string {
	return c.commits
}
