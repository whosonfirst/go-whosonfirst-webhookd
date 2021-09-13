CWD=$(shell pwd)

cli:
	go build -mod vendor -o bin/webhookd-flatten-config cmd/webhookd-flatten-config/main.go
	go build -mod vendor -o bin/webhookd-generate-hook cmd/webhookd-generate-hook/main.go
	go build -mod vendor -o bin/dispatch-buffered cmd/dispatch-buffered/main.go
	go build -mod vendor -o bin/launch-ecs-task cmd/launch-ecs-task/main.go

debug:
	# if test !-d /tmp/webhookd; then mkdir /tmp/webhookd; fi
	# if test !-d /tmp/findingaid; them mkdir /tmp/findingaid; fi
	go run -mod vendor cmd/webhookd/main.go -config-uri 'file://$(CWD)/docs/config/config.json.example?decoder=string'

debug-post:
	curl -v 'http://localhost:8080/insecure-test?debug=1' -d @docs/events/flights.json

debug-findingaid:
	curl -v 'http://localhost:8080/findingaid-test?debug=1' -d @docs/events/flights.json

lambda-config:
	go run cmd/webhookd-flatten-config/main.go -config $(CONFIG) -constvar | pbcopy

lambda:
	@make lambda-webhookd
	@make lambda-task

lambda-webhookd:
	if test -f main; then rm -f main; fi
	if test -f webhookd.zip; then rm -f webhookd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/webhookd/main.go
	zip webhookd.zip main
	rm -f main

lambda-task:
	if test -f main; then rm -f main; fi
	if test -f launch-ecs-task.zip; then rm -f launch-ecs-task.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/launch-ecs-task/main.go
	zip launch-ecs-task.zip main
	rm -f main
