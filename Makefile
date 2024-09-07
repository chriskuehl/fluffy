.PHONY: minimal
minimal: bin/server bin/fpb bin/fput assets

.PHONY: bin/server
bin/server:
	go build -o $@ ./cmd/server

.PHONY: bin/fpb
bin/fpb:
	go build -o $@ ./cmd/fpb

.PHONY: bin/fput
bin/fput:
	go build -o $@ ./cmd/fput

.PHONY: dev
dev:
	go run github.com/cespare/reflex@latest -v -s -r '^server/|^go\.mod$$' -- go run ./cmd/server --dev

.PHONY: delve
delve:
	dlv debug ./cmd/server -- --dev

.PHONY: release-cli
release-cli: export GORELEASER_CURRENT_TAG ?= 0.0.0
release-cli: export VERSION ?= 0.0.0
release-cli:
	go run github.com/goreleaser/goreleaser/v2@latest release --config .goreleaser-cli.yaml --clean --snapshot --verbose
	rm -v dist/*.txt dist/*.yaml dist/*.json

.PHONY: release-server
release-server: export GORELEASER_CURRENT_TAG ?= 0.0.0
release-server: export VERSION ?= 0.0.0
release-server:
	go run github.com/goreleaser/goreleaser/v2@latest release --clean --snapshot --verbose
	rm -v dist/*.txt dist/*.yaml dist/*.json

server/static/app.css: $(wildcard scss/*.scss)
	sass scss/app.scss $@

.PHONY: assets
assets: server/assets/static/app.css

.PHONY: watch-assets
watch-assets:
	sass --watch scss/app.scss:server/assets/static/app.css

.PHONY: test
test:
	go test ./...
