VENV := venv
BIN := $(VENV)/bin

.PHONY: all
all: $(VENV)

$(VENV): setup.py requirements.txt requirements-dev.txt
	vendor/venv-update venv= -p python3.4 venv install= -r requirements.txt -r requirements-dev.txt

.PHONY: test
test: $(VENV)
	$(BIN)/pre-commit install -f --install-hooks
	$(BIN)/pre-commit run --all-files

.PHONY: gunicorn
gunicorn: $(VENV)
	$(BIN)/gunicorn -w1 -b 0.0.0.0:6789 linux_wheels.upload_handler.app:app

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
