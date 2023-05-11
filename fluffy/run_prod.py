# This entrypoint is used to warm up the app at start.
print('Importing app...')
from fluffy.run import app  # noqa: E402 F401

# TODO: Consider removing this in a major version now that there is no warmup
# needed again.

print('Warmup complete')
