# go-whosonfirst-webhookd

A Who's On First specific instance of go-webhookd, along with related tools.

## go-webhookd

Before you begin please [read the go-webhookd documentation](https://github.com/whosonfirst/go-webhookd/blob/master/README.md) for an overview of concepts and principles.

## Usage

```
import (
	_ "github.com/go-whosonfirst-webhookd"
)
```

## Configuration

Please consult the [go-webhook configuration documentation](https://github.com/whosonfirst/go-webhookd#config-files) as well as the [example config file](docs/config/config.json.example).

## Tools

### webhookd

```
$> bin/webhookd -h
  -config-uri string
    	A valid Go Cloud runtimevar URI representing your webhookd config.
```

This build of the `webhookd` binary is the same as the tool defined in [whosonfirst/go-webhookd](https://github.com/whosonfirst/go-webhookd#webhookd) but imports the following packages:

```
import (
	// necessary for blob dispatcher and the findingaid-repo dispatcher (by way of the go-whosonfirst-findingaid cache)
	_ "github.com/aaronland/gocloud-blob-s3"                  
	// necessary for blob dispatcher	
	_ "gocloud.dev/blob/fileblob"                             
)

import (
	// defines the lambda dispatcher
	_ "github.com/whosonfirst/go-webhookd-aws/v2"
	// defines the github* transformations	
	_ "github.com/whosonfirst/go-webhookd-github"
	// defines the blob dispatcher	
	_ "github.com/whosonfirst/go-webhookd-gocloud"
)
```

#### AWS (Lambda)

```
$> make lambda-webhookd
if test -f main; then rm -f main; fi
if test -f webhookd.zip; then rm -f webhookd.zip; fi
GOOS=linux go build -mod vendor -o main cmd/webhookd/main.go
zip webhookd.zip main
  adding: main (deflated 59%)
rm -f main
```

_Documentation incomplete._

### dispatch-buffered

```
$> ./bin/dispatch-buffered -h
  -bucket-uri string
    	A valid gocloud.dev/blob Bucket URI where buffered dispatch messages are stored.
  -dryrun
    	Go through the motions but don't invoke any tasks
  -lambda-uri WEBHOOKD_LAMBDA_URIS
    	One or more valid aaronland/go-aws-lambda URIs. If run in -mode lambda mulitple values can be specified as a ';' separater list in the WEBHOOKD_LAMBDA_URIS environment variable.
  -mode string
    	Valid options are: cli, lambda. (default "cli")
```

The `dispatch-buffered` tool is meant to be used in conjunction with the `go-webhookd-github#repo` transformation and the `go-webhookd-gocloud#blob` dispatcher which will cause the body of the transformation (a WOF repo name) to be stored in an S3 bucket. The `dispatch-buffered`	tool will iterate over the files in the S3 bucket and invoke one or more Lambda functions with the body of each file (a WOF repo name).

This tool was written to account for (WOF) repos that receive a large number of atomic updates in a short amount of time (for example [sfomuseum-data/sfomuseum-data-media-collection](#)) and that trigger an unnecessarily large number of ECS tasks to be invoked (see "How it all fits together" below). The idea is that this tool can be run on a scheduler (CloudWatch, cron, whatever) and only process repos periodically.

```
#> ./bin/dispatch-buffered \
	-bucket-uri 's3blob://{BUCKET}/?region={REGION}&prefix={PREFIX}/&credentials=session' \
	-lambda-uri 'lambda://{LAMBDA}?dsn=region=us-west-2 credentials=session'
```

Notes:

* The `{LAMBDA}` function can be anything you want but in this example is assumed to be `cmd/launch-ecs-task`.

* By default, the `dispatch-buffered` tool is designed to trigger a Lambda function whose only purpose is to start an ECS task rather than starting that ECS task itself. While this is technically an extra step and another layer of indirection it's done this way to mimic, as closely as possible, the pattern that non-buffered webhook messages are processed with. Basically we're just buffering the "dispatch" layer which only knows to "pass this repo to a Lambda function" and doesn't know anything about ECS.

#### AWS (Lambda)

```
$> make lambda-dispatch
if test -f main; then rm -f main; fi
if test -f dispatch-buffered.zip; then rm -f dispatch-buffered.zip; fi
GOOS=linux go build -mod vendor -o main cmd/dispatch-buffered/main.go
zip dispatch-buffered.zip main
  adding: main (deflated 57%)
rm -f main
```

##### Environment variables

Your Lambda functions will need to following environment variables:

| Name | Value | Notes |
| --- | --- | --- |
| WEBHOOKD_BUCKET_URI	| `s3blob://{BUCKET}/?region={REGION}&prefix={PREFIX}/&credentials=iam:` | |
| WEBHOOKD_MODE	| `lambda` | |
| WEBHOOKD_DRYRUN | | true or false (default)

##### Policies

Your Lambda function will need to add the following IAM policies (or equivalents):

###### S3PendingBuffered

This allows the `dispatch-buffered` Lambda function to read (and delete) files (containing WOF repo names) to process.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetBucketLocation"
            ],
            "Resource": [
                "arn:aws:s3:::{BUCKET}"
            ]
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObjectAcl",
                "s3:GetObject",
                "s3:DeleteObject",
                "s3:PutObjectAcl"
            ],
            "Resource": [
                "arn:aws:s3:::{BUCKET}/buffered/*"
            ]
        },
        {
            "Sid": "VisualEditor2",
            "Effect": "Allow",
            "Action": "s3:ListAllMyBuckets",
            "Resource": [
                "arn:aws:s3:::{BUCKET}/buffered/*"
            ]
        }
    ]
}
```

###### LambdaDispatchBuffered

This allows the `dispatch-buffered` Lambda function to invoke one (or more) _other_ Lambda functions.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "lambda:InvokeFunction",
            "Resource": [
	    	"arn:aws:lambda:{REGION}:{ACCOUNT_ID}:function:{LAMBDA_FUNCTION}",
		"arn:aws:lambda:{REGION}:{ACCOUNT_ID}:function:{LAMBDA_FUNCTION}"
	    ]
        }
    ]
}
```

##### Role

Your Lambda function will need an IAM role with the following policies:

* `AWSLambdaExecute`
* `S3PendingBuffered`
* `LambdaDispatchBuffered`

### launch-ecs-task

The `launch-ecs-task` tool, as you might expect, launches an ECS task. This is actually a pretty generic tool that might be better kept in another package. There is nothing Who's On First specific about this tool but in as much as it was built to be part of the larger Who's On First specific data processing pipeline (see "How it all fits together" below) we're going to keep it in this package for the time being.

```
$> ./bin/launch-ecs-task -h
  -command string
    	The command to launch your ECS task with.
  -ecs-cluster string
    	The name of your AWS ECS cluster.
  -ecs-container string
    	The name of your AWS ECS container.
  -ecs-dsn string
    	A valid aaronland/go-aws-ecs DSN string.
  -ecs-launch-type string
    	A valid (AWS) ECS launch type. (default "FARGATE")
  -ecs-public-ip string
    	Whether or not to enable a public IP address for your ECS task. (default "ENABLED")
  -ecs-security-group value
    	One of more AWS security groups your task will assume.
  -ecs-subnet value
    	One or more AWS subnets in which your task will run.
  -ecs-task string
    	The name of your AWS ECS task (inclusive of its version number),
  -mode string
    	Valid options are: cli, lambda. (default "cli")
```

#### AWS (Lambda)

```
$> make lambda-task
if test -f main; then rm -f main; fi
if test -f launch-ecs-task.zip; then rm -f launch-ecs-task.zip; fi
GOOS=linux go build -mod vendor -o main cmd/launch-ecs-task/main.go
zip launch-ecs-task.zip main
  adding: main (deflated 54%)
rm -f main
```

##### Enivronment variables

Your Lambda functions will need to following environment variables:

| Name | Value | Notes |
| --- | --- | --- |
| WEBHOOKD_COMMAND | | The command to invoke your ECS task with |
| WEBHOOKD_ECS_CLUSTER	| | The name of the ECS cluster for your task |
| WEBHOOKD_ECS_CONTAINER | | The name of your ECS cluster | 
| WEBHOOKD_ECS_DSN | | A valid `aaronland/go-aws-ecs` DSN string | 
| WEBHOOKD_ECS_SECURITY_GROUP | | The (AWS) security group to use for your ECS task | 
| WEBHOOKD_ECS_SUBNET | | The (AWS) subnet to use for your ECS task | 
| WEBHOOKD_ECS_TASK | | The name and version of the ECS task to launch (for example `sfomuseum-data-indexing:45`)
| WEBHOOKD_MODE	| `lambda` | | 

###### Policies (IAM)

Your Lambda function will need to add the following IAM policies (or equivalents):

###### ECSLaunchTask

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Stmt1",
            "Effect": "Allow",
            "Action": [
                "ecs:RunTask"
            ],
            "Resource": [
                "arn:aws:ecs:{REGION}:{ACCOUNT_ID}:task-definition/{TASK_DEFINITION}:*"
            ]
        },
        {
            "Sid": "Stmt2",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole"
            ],
            "Resource": [
                "arn:aws:iam::{ACCOUNT_ID}:role/ecsTaskExecutionRole",
                "arn:aws:iam::{ACCOUNT_ID}:role/ECSLaunchTask"
            ]
        }
    ]
}
```

###### Roles (IAM)

Your Lambda function will need an IAM role with the following policies:

* AWSLambdaExecute
* ECSLaunchTask

### webhookd-generate-hook

```
$> ./bin/webhookd-generate-hook -h
Usage of ./bin/webhookd-generate-hook:
  -length int
    	The length of your webhook (default 64)
```

`webhookd-generate-hook` will generate a random string to use as a webhook endpoint.

### webhookd-flatten-config

```
> ./bin/webhookd-flatten-config -h
Usage of ./bin/webhookd-flatten-config:
  -config string
    	The path your webhookd config file.
  -constvar constvar
    	Encode the output as a valid gocloud.dev/runtimevar constvar string.
```

`webhookd-flatten-config` is a tool to "flatten" a valid [webhookd config file](https://github.com/whosonfirst/go-webhookd#config-files) in to a string that can be copy-pasted in to an (AWS) Lambda environment variable field.

## See also

* https://github.com/whosonfirst/go-webhookd
* https://github.com/whosonfirst/go-webhookd-aws
* https://github.com/whosonfirst/go-webhookd-github
* https://github.com/whosonfirst/go-webhookd-gocloud