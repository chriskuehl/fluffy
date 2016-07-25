VENV := venv
BIN := $(VENV)/bin
DOCKER_TAG := ckuehl/fluffy-server

export FLUFFY_SETTINGS := $(PWD)/settings.py

.PHONY: minimal
minimal: $(VENV) assets settings.py

$(VENV): setup.py cli/setup.py requirements.txt requirements-dev.txt
	vendor/venv-update venv= -p python3.4 venv install= -r requirements.txt -r requirements-dev.txt

fluffy/static/app.css: $(VENV) $(wildcard fluffy/static/scss/*.scss)
	$(BIN)/sassc fluffy/static/scss/app.scss $@


fluffy/static/js/icons.js: $(VENV) fluffy/static/img/mime/small fluffy/assets.py settings.py
	$(BIN)/fluffy-build-icons-js > $@

fluffy/static/js/icons.debug.js: $(VENV) fluffy/static/img/mime/small fluffy/assets.py settings.py
	$(BIN)/fluffy-build-icons-js-debug > $@


ASSET_FILES := $(shell find fluffy/static -type f -not -name '*.hash')
ASSET_HASHES := $(addsuffix .hash,$(ASSET_FILES))

fluffy/static/%.hash: fluffy/static/%
	sha256sum $^ | awk '{print $$1}' > $@

.PHONY: assets
assets: fluffy/static/app.css.hash fluffy/static/js/icons.debug.js.hash fluffy/static/js/icons.js.hash $(ASSET_HASHES)

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
	$(BIN)/coverage run $(BIN)/py.test -vv tests/
	$(BIN)/coverage combine
	$(BIN)/coverage report
	$(BIN)/pre-commit install -f --install-hooks
	$(BIN)/pre-commit run --all-files

.PHONY: docker-image
docker-image: assets
	docker build -t $(DOCKER_TAG) .
	@echo 'Maybe you want to run:'
	@echo -e '    \033[1mdocker push ckuehl/fluffy-server\033[0m'

.PHONY: docker-run
docker-run: docker-image
	docker run -p 8000 $(DOCKER_TAG)

.PHONY: clean
clean:
	rm -rf $(VENV)

.PHONY: release
release: $(VENV)
	# server
	$(BIN)/python setup.py sdist
	$(BIN)/python setup.py bdist_wheel
	$(BIN)/twine upload --skip-existing dist/*
	# cli
	cd cli && ../$(BIN)/python setup.py sdist
	cd cli && ../$(BIN)/python setup.py bdist_wheel
	cd cli && ../$(BIN)/twine upload --skip-existing dist/*

.PHONY: update-requirements
update-requirements:
	$(eval TMP := $(shell mktemp -d))
	python ./vendor/venv-update venv= $(TMP) -ppython3 install= .
	. $(TMP)/bin/activate && \
		pip freeze | sort | grep -vE '^(wheel|pip-faster|virtualenv|fluffy-server)==' > requirements.txt
	rm -rf $(TMP)
