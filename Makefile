VENV := venv
BIN := $(VENV)/bin
DOCKER_TAG := ckuehl/fluffy-server

export FLUFFY_SETTINGS := $(PWD)/settings.py

.PHONY: all
all: $(VENV)

$(VENV): setup.py requirements.txt requirements-dev.txt
	vendor/venv-update venv= -p python3.4 venv install= -r requirements.txt -r requirements-dev.txt

.PHONY: dev
dev: $(VENV)
	$(BIN)/python -m fluffy.run

.PHONY: test
test: $(VENV)
	$(BIN)/pre-commit install -f --install-hooks
	$(BIN)/pre-commit run --all-files

.PHONY: docker-image
docker-image:
	docker build -t $(DOCKER_TAG) .

.PHONY: docker-run
docker-run: docker-image
	docker run -p 8000 $(DOCKER_TAG)

.PHONY: clean
clean:
	rm -rf $(VENV)

.PHONY: update-requirements
update-requirements:
	$(eval TMP := $(shell mktemp -d))
	python ./vendor/venv-update venv= $(TMP) -ppython3 install= .
	. $(TMP)/bin/activate && \
		pip freeze | sort | grep -vE '^(wheel|pip-faster|virtualenv|fluffy-server)==' > requirements.txt
	rm -rf $(TMP)
