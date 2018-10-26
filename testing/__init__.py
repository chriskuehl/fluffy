import re

import requests


PLAINTEXT_TESTCASES = (
    '',
    '\t\t\t',
    '    ',
    'hello world',
    'éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ',
    'hello\nworld\n',
)

BINARY_TESTCASES = (
    b'hello world\00',
    b'\x43\x92\xd9\x0f\xaf\x32\x2c\x00\x12\x23',
    b'\x11\x22\x33\x44\x55',
)

FILE_CONTENT_TESTCASES = tuple(
    content.encode('utf8')
    for content in PLAINTEXT_TESTCASES
) + BINARY_TESTCASES


def urls_from_details(details):
    """Return list of URLs to objects from details page source."""
    return re.findall(
        r'<a href="(http://localhost:\d+/object/[^"]+)"',
        details,
    )


def paste_urls_from_details(details):
    """Return list of URLs to objects from details page source."""
    return re.findall(
        r'<a href="(http://localhost:\d+/html/[^"]+)"',
        details,
    )


def raw_text_url_from_paste_html(paste_html):
    """Return raw text URL from a paste page source."""
    url, = re.findall(
        r'<a class="button" href="(http://localhost:\d+/object/[^"]+)">\s+Raw Text',
        paste_html,
    )
    return url


def assert_url_matches_content(url, content):
    req = requests.get(url)
    assert req.content == content
