# This entrypoint is used to warm up the app at start.
print('Importing app...')
from fluffy.run import app  # noqa: E402 F401

print('Initializing guesslang...')
import fluffy.component.highlighting  # noqa: E402
fluffy.component.highlighting._guesslang_guesser()

print('Warmup complete')
