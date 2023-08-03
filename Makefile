SHELL = /bin/bash
.SILENT: # comment this one out if you need to debug the Makefile
APP=logseq-export

.PHONY: build
build:
	go build ./...
	# go build -o ${APP} main.go

.PHONY: test
test:
	go test ./...

.PHONY: watch-test
watch-test:
	fswatch \
		--exclude '.git' \
		--exclude 'logseq-export$$' \
		--exclude 'test/test-output' \
		--exclude 'example/' \
		-o . | \
		xargs -n1 -I{} go test ./...

.PHONY: clean
clean:
	go clean

.PHONY: example
example:
	$(MAKE) build
	tmp_export_folder=$$(mktemp -d); \
	./logseq-export \
		--logseqFolder "$(CURDIR)/example/logseq-graph" \
		--outputFolder "$$tmp_export_folder"; \
	./import_to_hugo.sh "$$tmp_export_folder" "$(CURDIR)/example/logseq-export-example"

.PHONY: watch-example
watch-example:
	fswatch \
		--exclude '.git' \
		--exclude 'logseq-export$$' \
		--exclude 'test/test-output' \
		--exclude 'example/logseq-export-example/content/graph' \
		--exclude 'example/logseq-export-example/static/assets/graph' \
		-xn . | \
		while read file event; do \
			echo "File $$file has changed, Event: $$event"; \
			$(MAKE) example; \
		done
