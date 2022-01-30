.PHONY: build clean deploy

build:
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o out/fetch fetch/cmd/main.go

build-docker:
	docker run -v "$(shell pwd)":/go/src/handler -e GOARCH=amd64 -e GOOS=linux lambci/lambda:build-go1.x go build -ldflags="-s -w" -o out/fetch fetch/cmd/main.go

test:
	go test -race -coverprofile out/coverage.out ...

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose

test-locally:
	sls invoke local -f fetch -p testdata/event.json -s local

start-localstack: deps
	localstack start -d

deps:
	pip install localstack

gen-mocks:
	mockery --all --output fetch/mocks