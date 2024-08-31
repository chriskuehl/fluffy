VENV := venv
BIN := $(VENV)/bin
export FLUFFY_SETTINGS := $(CURDIR)/settings.py

.PHONY: minimal
minimal: $(VENV) bin/server bin/fpb bin/fput assets settings.py install-hooks

.PHONY: fluffy-server
bin/server:
	go build -o $@ ./cmd/server

.PHONY: dev
dev:
	go run ./cmd/server


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

$(VENV): setup.py requirements.txt requirements-dev.txt
	rm -rf $@
	virtualenv -ppython3.11 $@
	$@/bin/pip install -r requirements.txt -r requirements-dev.txt -e .
	ln -fs ../../bin/fput $@/bin/fput
	ln -fs ../../bin/fpb $@/bin/fpb

fluffy/static/app.css: $(VENV) $(wildcard fluffy/static/scss/*.scss)
	$(BIN)/pysassc fluffy/static/scss/app.scss $@

fluffy/static/pygments.css: $(VENV) fluffy/component/styles.py
	$(BIN)/python -m fluffy.component.styles > $@

ASSET_FILES := $(shell find fluffy/static -type f -not -name '*.hash')
ASSET_HASHES := $(addsuffix .hash,$(ASSET_FILES))

fluffy/static/%.hash: fluffy/static/%
	sha256sum $^ | awk '{print $$1}' > $@

.PHONY: assets
assets: fluffy/static/app.css.hash fluffy/static/pygments.css.hash $(ASSET_HASHES)

.PHONY: upload-assets
upload-assets: assets $(VENV)
	$(BIN)/fluffy-upload-assets

settings.py:
	ln -fs settings/dev_files.py settings.py

.PHONY: watch-assets
watch-assets:
	while :; do \
		find fluffy/static -type f | \
			inotifywait --fromfile - -e modify; \
			make assets; \
	done

.PHONY: test
test: $(VENV)
	go test -v ./...
	$(BIN)/coverage erase
	COVERAGE_PROCESS_START=$(CURDIR)/.coveragerc \
		$(BIN)/py.test --tb=native -vv tests/
	$(BIN)/coverage combine
	$(BIN)/coverage report

.PHONY: install-hooks
install-hooks: $(VENV)
	$(BIN)/pre-commit install -f --install-hooks

.PHONY: pre-commit
pre-commit: $(VENV)
	$(BIN)/pre-commit run --all-files

.PHONY: clean
clean:
	rm -rf $(VENV)

.PHONY: upgrade-requirements
upgrade-requirements: venv
	upgrade-requirements
