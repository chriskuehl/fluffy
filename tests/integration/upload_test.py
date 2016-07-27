import io
import re

import mock
import pytest
import requests


FILE_CONTENT_TESTCASES = (
    b'',
    b'hello world',
    b'hello\nworld\n',
    'éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ'.encode('utf8'),
    '\x43\x92\xd9\x0f\xaf\x32\x2c\x00\x12\x23'.encode('utf8'),
)


def _urls_from_details(details):
    """Return list of URLs to objects from details page source."""
    return re.findall(
        r'<a href="(http://localhost:\d+/object/[^"]+\.bin)"',
        details,
    )


def _assert_url_matches_content(url, content):
    req = requests.get(url)
    assert req.content == content


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
def test_single_file_upload(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload',
        files=[('file', ('ohai.bin', io.BytesIO(content), None, None))],
    )
    assert req.status_code == 200
    assert 'ohai.bin' in req.text

    url, = _urls_from_details(req.text)
    _assert_url_matches_content(url, content)


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
def test_single_file_upload_json(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload?json',
        files=[('file', ('ohai.bin', io.BytesIO(content), None, None))],
    )
    assert req.status_code == 200
    assert req.json() == {'success': True, 'redirect': mock.ANY}

    req = requests.get(req.json()['redirect'])
    assert req.status_code == 200
    url, = _urls_from_details(req.text)
    _assert_url_matches_content(url, content)


def test_multiple_files_upload(running_server):
    files = [
        ('file', ('ohai{}.bin'.format(i), io.BytesIO(content), None, None))
        for i, content in enumerate(FILE_CONTENT_TESTCASES)
    ]
    req = requests.post(
        running_server['home'] + '/upload',
        files=files,
    )
    assert req.status_code == 200
    urls = _urls_from_details(req.text)
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        assert 'ohai{}.bin'.format(i) in req.text
        _assert_url_matches_content(urls[i], content)


def test_multiple_files_upload_json(running_server):
    files = [
        ('file', ('ohai{}.bin'.format(i), io.BytesIO(content), None, None))
        for i, content in enumerate(FILE_CONTENT_TESTCASES)
    ]
    req = requests.post(
        running_server['home'] + '/upload?json',
        files=files,
    )
    assert req.status_code == 200
    assert req.json() == {'success': True, 'redirect': mock.ANY}

    req = requests.get(req.json()['redirect'])
    assert req.status_code == 200

    urls = _urls_from_details(req.text)
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        assert 'ohai{}.bin'.format(i) in req.text
        _assert_url_matches_content(urls[i], content)
