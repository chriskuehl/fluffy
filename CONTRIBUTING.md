## Release process (server)

1. Bump version in `fluffy/__init__.py`
2. Commit, tag as "vX.Y.Z", and push to main.
3. A GitHub Actions workflow will build and publish to PyPI.


## Release process (cli)

1. Bump version in `cli/fluffy_cli/__init__.py`
2. cd to `cli` and use `dch` to update the Debian changelog
3. Run `make release` and upload a new deb to GitHub releases.
