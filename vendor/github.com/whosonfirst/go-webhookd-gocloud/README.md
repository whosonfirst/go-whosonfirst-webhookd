# go-webhookd-gocloud

Go package to implement the `whosonfirst/go-webhookd` interfaces for dispatching messages to gocloud.dev/blob resources.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-webhookd-gocloud.svg)](https://pkg.go.dev/github.com/whosonfirst/go-webhookd-gocloud)

### go-webhookd

Before you begin please [read the go-webhookd documentation](https://github.com/whosonfirst/go-webhookd/blob/master/README.md) for an overview of concepts and principles.

## Usage

```
import (
	_ "github.com/whosonfirst/go-webhookd-gocloud"
)
```

## Example

```
import (
	_ "gocloud.dev/blob/memblob"
)

import (
	"context"
	"github.com/whosonfirst/go-webhookd/v3/dispatcher"
	_ "github.com/whosonfirst/go-webhookd-gocloud"
)

func main()

	ctx := context.Background()          
	d, _ := dispatcher.NewDispatcher(ctx, "mem://")
	
	d.Dispatch(ctx, []byte("hello world"))
}	
```

_Error handling omitted for the sake of brevity._

## See also

* https://github.com/whosonfirst/go-webhookd
* https://gocloud.dev/