VENV := venv
BIN := $(VENV)/bin
export FLUFFY_SETTINGS := $(CURDIR)/settings.py

.PHONY: minimal
minimal: $(VENV) assets settings.py install-hooks

$(VENV): setup.py cli/setup.py requirements.txt requirements-dev.txt
	rm -rf $@
	virtualenv -ppython3.11 $@
	$@/bin/pip install -r requirements.txt -r requirements-dev.txt -e cli -e .

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

.PHONY: dev
dev: $(VENV) fluffy/static/app.css
	$(BIN)/python -m fluffy.run

.PHONY: test
test: $(VENV)
	$(BIN)/coverage erase
	COVERAGE_PROCESS_START=$(CURDIR)/.coveragerc \
		$(BIN)/py.test -vv tests/
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
