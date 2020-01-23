.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/auth auth/*.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/telegram telegram/*.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
