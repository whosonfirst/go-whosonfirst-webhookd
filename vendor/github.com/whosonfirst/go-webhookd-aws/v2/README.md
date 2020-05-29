# go-webhookd-aws

go-webhookd support for Amazon Web Services (AWS)

## go-webhookd

Before you begin please [read the go-webhookd documentation](https://github.com/whosonfirst/go-webhookd/blob/master/README.md) for an overview of concepts and principles.

## Important

`whosonfirst/go-webhookd-aws/v2` is backwards incompatible with `whosonfirst/go-webhookd-aws` "v1". Importantly the ability to run a `webhookd` server _as_ an AWS Lambda has been merged back in to `whosonfirst/go-webhookd/v2` (and higher). This package only manages AWS specific dispatchers now.

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

## See also

* https://github.com/whosonfirst/go-webhookd