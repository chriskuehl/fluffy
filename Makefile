.PHONY: minimal
minimal: bin/server bin/fpb bin/fput assets settings.py install-hooks

.PHONY: fluffy-server
bin/server:
	go build -o $@ ./cmd/server

.PHONY: dev
dev:
	go run github.com/cespare/reflex@latest -v -s -r '^server/|^go\.mod$$' -- go run ./cmd/server --dev

.PHONY: bin/fpb
bin/fpb:
	go build -o $@ ./cli/fpb

.PHONY: bin/fput
bin/fput:
	go build -o $@ ./cli/fput

.PHONY: release-cli
release-cli: export GORELEASER_CURRENT_TAG ?= 0.0.0
release-cli: export VERSION ?= 0.0.0
release-cli:
	go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot --verbose
	rm -v dist/*.txt dist/*.yaml dist/*.json

server/static/app.css: $(wildcard scss/*.scss)
	sass scss/app.scss $@

.PHONY: assets
assets: server/static/app.css

.PHONY: watch-assets
watch-assets:
	sass --watch scss/app.scss:server/static/app.css

.PHONY: test
test:
	go test -v ./...
