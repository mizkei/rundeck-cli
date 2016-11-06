all: rundeck-cli

get-deps:
	go get ./...

test:
	go test -v ./...

.PHONY: rundeck-cli
rundeck-cli:
	go build -o rundeck-cli main.go conf.go completion.go
