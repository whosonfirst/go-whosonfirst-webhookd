lambda: lambda-webhookd lambda-task

lambda-webhookd:
	if test -f main; then rm -f main; fi
	if test -f webhookd-aws.zip; then rm -f webhookd-aws.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd-aws/main.go
	zip webhookd-aws.zip main
	rm -f main

lambda-task:
	if test -f main; then rm -f main; fi
	if test -f webhookd-aws-task.zip; then rm -f webhookd-aws-task.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd-aws-launch-task/main.go
	zip webhookd-aws-task.zip main
	rm -f main
