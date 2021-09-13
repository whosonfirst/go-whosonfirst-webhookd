lambda: lambda-webhookd

lambda-webhookd:
	if test -f main; then rm -f main; fi
	if test -f webhookd-aws.zip; then rm -f webhookd-aws.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd-aws/main.go
	zip webhookd-aws.zip main
	rm -f main
