lambda-config:
	go run cmd/webhookd-flatten-config/main.go -config $(CONFIG) -constvar | pbcopy

lambda: lambda-webhookd

lambda-webhookd:
	if test -f main; then rm -f main; fi
	if test -f webhookd.zip; then rm -f webhookd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd/main.go
	zip webhookd.zip main
	rm -f main

