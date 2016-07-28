import re

import requests


FILE_CONTENT_TESTCASES = (
    b'',
    b'hello world',
    b'hello\nworld\n',
    'éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ'.encode('utf8'),
    '\x43\x92\xd9\x0f\xaf\x32\x2c\x00\x12\x23'.encode('utf8'),
)


def urls_from_details(details):
    """Return list of URLs to objects from details page source."""
    return re.findall(
        r'<a href="(http://localhost:\d+/object/[^"]+\.bin)"',
        details,
    )


def assert_url_matches_content(url, content):
    req = requests.get(url)
    assert req.content == content
