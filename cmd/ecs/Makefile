.PHONY: build clean deploy

build:
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o out/fetch fetch/cmd/main.go

build-docker:
	docker run -v "$(shell pwd)/../../":/go/src/handler -e GOARCH=amd64 -e GOOS=linux lambci/lambda:build-go1.x go build -ldflags="-s -w" -o cmd/ecs/out/fetch cmd/ecs/main.go
	chmod +x ./out/fetch

test:
	go test -race -coverprofile out/coverage.out ...

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose

test-locally: start-localstack build-docker 
	sls invoke local -f fetch --docker-arg "--network ecs_localstack" -p testdata/event.json -s local

start-localstack: 
	docker-compose up -d

gen-mocks:
	mockery --all --output fetch/mocks