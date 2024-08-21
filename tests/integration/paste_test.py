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


def test_simple_paste_diff_between_two_texts(running_server):
    req = requests.post(
        running_server['home'] + '/paste',
        data={
            'diff1': (
                'line A\n'
                'line B\n'
                'line C\n'
            ),
            'diff2': (
                'line B\n'
                'line B2\n'
                'line C\n'
                'line D\n'
            ),
            'language': 'diff-between-two-texts',
        },
    )

    assert req.status_code == 200
    pqq = pq(req.content.decode('utf8'))

    assert (
        pqq.find('input[name=text]').attr('value') ==
        '--- \n'
        '+++ \n'
        '@@ -1,3 +1,4 @@\n'
        '-line A\n'
        ' line B\n'
        '+line B2\n'
        ' line C\n'
        '+line D'
    )

    assert ' -line A' in pq(pqq.find('.text')[0]).text()
    assert ' -line A' not in pq(pqq.find('.text')[1]).text()

    assert ' +line B2' not in pq(pqq.find('.text')[0]).text()
    assert ' +line B2' in pq(pqq.find('.text')[1]).text()


def test_simple_paste_json(running_server):
    req = requests.post(
        running_server['home'] + '/paste?json',
        data={
            'text': 'hello world',
            'language': 'python',
        },
    )

    assert req.status_code == 200
    assert req.headers['Content-Type'] == 'application/json'
    assert req.json() == {
        'success': True,
        'redirect': mock.ANY,
        'metadata': mock.ANY,
        'uploaded_files': {
            'paste': {
                'raw': mock.ANY,
                'paste': mock.ANY,
                'num_lines': 1,
                'language': {
                    'title': 'Python',
                },
            },
        },
    }

    # The paste's HTML view and raw view should have the same URL minus the extension.
    details = req.json()['uploaded_files']['paste']
    assert details['raw'].rsplit('/', 1)[1] == details['paste'].replace('.html', '.txt').rsplit('/', 1)[1]
