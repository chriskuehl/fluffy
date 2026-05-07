import hashlib
from pathlib import Path

import requests

PROJECT_ROOT = Path(__file__).parent.parent.parent


def test_jquery_sha256():
    """Verify the vendored jQuery matches the file served by the CDN."""

    data = (
        PROJECT_ROOT / 'fluffy' / 'static' / 'vendor' / 'js' / 'jquery.min.js'
    ).read_bytes()

    cdn = requests.get(
        'https://ajax.googleapis.com/ajax/libs/jquery/3.1.0/jquery.min.js',
    )
    cdn.raise_for_status()

    # SHA256: 702b9e051e82b32038ffdb33a4f7eb5f7b38f4cf6f514e4182d8898f4eb0b7fb
    assert hashlib.sha256(data).hexdigest() == hashlib.sha256(cdn.content).hexdigest()
