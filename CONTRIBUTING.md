## Development setup

Prerequisites: Python 3.14+, Go 1.23+, [uv](https://docs.astral.sh/uv/).

```bash
make minimal          # creates venv, builds fpb/fput, compiles assets, symlinks settings.py
make dev              # starts the Flask dev server on port 5000
```

You also need a static file server to serve uploaded content locally:

```bash
mkdir -p tmp/object tmp/html
python -m http.server 4999 --directory tmp
```

## Running tests

`make minimal` must complete before tests will pass (assets and Go binaries are required).

```bash
make test             # Go tests + Python tests with coverage
```


## Release process (server)

1. Bump version in `fluffy/__init__.py`
2. Commit, tag as "vX.Y.Z", and push to main.
3. A GitHub Actions workflow will build and publish to PyPI.


## Release process (cli)

1. Bump version in `cli/fluffy_cli/__init__.py`
2. cd to `cli` and use `dch` to update the Debian changelog
3. Run `make release` and upload a new deb to GitHub releases.
