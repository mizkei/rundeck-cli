all: go-rundeck

get-deps:
		go get ./...

go-rundeck:
		go build -o rundeck-cli main.go conf.go
