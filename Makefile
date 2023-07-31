APP=logseq-export

.PHONY: build
build:
	go build ./...
	# go build -o ${APP} main.go

.PHONY: test
test:
	go test ./...

.PHONY: watch
watch:
	fswatch --exclude 'test/test-output' -o ./ | xargs -n1 -I{} go test ./...

.PHONY: clean
clean:
	go clean
