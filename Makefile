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
	go run github.com/cespare/reflex@latest -v -s -r '^server/|^go\.mod$$' -- go run -race ./cmd/server --dev

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
	sass --style compressed scss/app.scss $@

.PHONY: server/static/chroma.css
server/static/chroma.css:
	go run ./cmd/print-styles-css | sass --stdin --style compressed $@

.PHONY: assets
assets: server/static/app.css server/static/chroma.css

.PHONY: watch-assets
watch-assets:
	sass --style compressed --watch scss/app.scss:server/static/app.css

.PHONY: test
test:
	go test -coverprofile cover.out -race ./...
	go tool cover -html=cover.out -o cover.html
