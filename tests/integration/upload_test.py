import io

import mock
import pytest
import requests
from pyquery import PyQuery as pq

from testing import assert_url_matches_content
from testing import BINARY_TESTCASES
from testing import FILE_CONTENT_TESTCASES
from testing import paste_urls_from_details
from testing import PLAINTEXT_TESTCASES
from testing import urls_from_details


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
def test_single_file_upload(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload',
        files=[('file', ('ohai.bin', io.BytesIO(content), None, None))],
    )
    assert req.status_code == 200
    assert 'ohai.bin' in req.text

    url, = urls_from_details(req.text)
    assert_url_matches_content(url, content)


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
    url, = urls_from_details(req.text)
    assert_url_matches_content(url, content)


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
    urls = urls_from_details(req.text)
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        assert 'ohai{}.bin'.format(i) in req.text
        assert_url_matches_content(urls[i], content)


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

    urls = urls_from_details(req.text)
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        assert 'ohai{}.bin'.format(i) in req.text
        assert_url_matches_content(urls[i], content)


@pytest.mark.parametrize('content', PLAINTEXT_TESTCASES)
def test_plaintext_files_are_also_pasted(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload',
        files=[('file', ('ohai.bin', io.StringIO(content), None, None))],
    )
    assert req.status_code == 200
    url, = paste_urls_from_details(req.text)

    req = requests.get(url)
    assert (
        pq(req.content.decode('utf8')).find('input[name=text]').attr('value') ==
        content
    )


@pytest.mark.parametrize('content', BINARY_TESTCASES)
def test_binary_files_are_not_pasted(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload',
        files=[('file', ('ohai.bin', io.BytesIO(content), None, None))],
    )
    assert req.status_code == 200
    assert paste_urls_from_details(req.text) == []
