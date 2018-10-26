from unittest import mock

import pytest
import requests
from pyquery import PyQuery as pq

from testing import PLAINTEXT_TESTCASES


@pytest.mark.parametrize('content', PLAINTEXT_TESTCASES)
def test_simple_paste(content, running_server):
    req = requests.post(
        running_server['home'] + '/paste',
        data={
            'text': content,
            'language': 'autodetect',
        },
    )

    assert req.status_code == 200
    assert (
        pq(req.content.decode('utf8')).find('input[name=text]').attr('value') ==
        content
    )


def test_simple_paste_json(running_server):
    req = requests.post(
        running_server['home'] + '/paste?json',
        data={
            'text': 'hello world',
            'language': 'autodetect',
        },
    )

    assert req.status_code == 200
    assert req.headers['Content-Type'] == 'application/json'
    assert req.json() == {
        'success': True,
        'redirect': mock.ANY,
        'uploaded_files': {
            'paste': {
                'raw': mock.ANY,
                'paste': mock.ANY,
                'metadata': mock.ANY,
            },
        },
    }
