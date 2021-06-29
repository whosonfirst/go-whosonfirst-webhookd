CWD=$(shell pwd)

debug:
	# if test !-d /tmp/webhookd; then mkdir /tmp/webhookd; fi
	# if test !-d /tmp/findingaid; them mkdir /tmp/findingaid; fi
	go run -mod vendor cmd/webhookd/main.go -config-uri 'file://$(CWD)/docs/config/config.json.example?decoder=string'

debug-post:
	curl -v 'http://localhost:8080/insecure-test?debug=1' -d @docs/events/flights.json

lambda-config:
	go run cmd/webhookd-flatten-config/main.go -config $(CONFIG) -constvar | pbcopy

lambda: lambda-webhookd

lambda-webhookd:
	if test -f main; then rm -f main; fi
	if test -f webhookd.zip; then rm -f webhookd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd/main.go
	zip webhookd.zip main
	rm -f main

