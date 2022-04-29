# go-aws-ecs

Go package for basic AWS ECS related operations.

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/ecs-launch-task cmd/ecs-launch-task/main.go
```

### ecs-launch-task

Launch an ECS task from the command line.

```
$> ./bin/ecs-launch-task -h
Usage of ./bin/ecs-launch-task:
  -cluster string
    	The name of your ECS cluster.
  -container string
    	The name of your ECS container.
  -dsn string
    	A valid aaronland/go-aws-session DSN string.
  -launch-type string
    	A valid ECS launch type.
  -platform-version string
    	A valid ECS platform version.
  -public-ip string
    	A valid ECS public IP string.
  -security-group value
    	A valid AWS security group to run your task under.
  -subnet value
    	One or more subnets to run your ECS task in.
  -task string
    	The name (and version) of your ECS task.
```

#### DSN strings

The following properties are required in DSN strings:

### Credentials

Credentials for AWS sessions are defined as string labels. They are:

| Label | Description |
| --- | --- |
| `env:` | Read credentials from AWS defined environment variables. |
| `iam:` | Assume AWS IAM credentials are in effect. |
| `{AWS_PROFILE_NAME}` | This this profile from the default AWS credentials location. |
| `{AWS_CREDENTIALS_PATH}:{AWS_PROFILE_NAME}` | This this profile from a user-defined AWS credentials location. |

For example:

```
region=us-east-1&credentials=session
```

### Region

Any valid AWS region.

## See also

* https://github.com/aws/aws-sdk-go
* https://github.com/aaronland/go-aws-session