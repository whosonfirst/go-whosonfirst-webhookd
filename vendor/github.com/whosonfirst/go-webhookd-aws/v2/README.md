# go-webhookd-aws

Go package to implement the `whosonfirst/go-webhookd` interfaces for dispatching webhooks originating from GitHub to AWS services.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-webhookd-aws.svg)](https://pkg.go.dev/github.com/whosonfirst/go-webhookd-aws)

Before you begin please [read the go-webhookd documentation](https://github.com/whosonfirst/go-webhookd/blob/master/README.md) for an overview of concepts and principles.

## Usage

```
import (
	_ "github.com/go-webhookd-aws/v2"
)
```

## Dispatchers

### Lambda

The `Lambda` dispatcher will send messages to an Amazon Web Services (ASW) [Lambda function](#). It is defined as a URI string in the form of:

```
lambda://{FUNCTION}?dsn={DSN}&invocation_type={INVOCATION_TYPE}
```

#### Properties

| Name | Value | Description | Required |
| --- | --- | --- | --- |
| dsn | string | A valid `aaronland/go-aws-session` DSN string. | yes |
| function | string | The name of your Lambda function. | yes |
| invocation_type | string | A valid AWS Lambda `Invocation Type` string. | no |
| halt_on_message | string | An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no |
| halt_on_author | string | An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no |

## Important

`whosonfirst/go-webhookd-aws/v2` and higher is backwards incompatible with `whosonfirst/go-webhookd-aws` "v1". Importantly the ability to run a `webhookd` server _as_ an AWS Lambda has been merged back in to `whosonfirst/go-webhookd/v2` (and higher). This package only manages AWS specific dispatchers now.


## See also

* https://github.com/whosonfirst/go-webhookd