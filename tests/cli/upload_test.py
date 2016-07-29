import os
import subprocess
import sys

import mock
import pytest
import requests

from testing import assert_url_matches_content
from testing import FILE_CONTENT_TESTCASES
from testing import urls_from_details


@pytest.yield_fixture
def cli_on_path():
    with mock.patch.dict(
            os.environ,
            {
                'PATH': '{}:{}'.format(
                    os.path.dirname(sys.executable), os.environ['PATH'],
                ),
            },
    ):
        yield


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
@pytest.mark.usefixtures('cli_on_path')
def test_single_file_upload(content, running_server, tmpdir):
    path = tmpdir.join('ohai.bin')
    path.write(content, 'wb')
    info_url = subprocess.check_output(
        ('fput', '--server', running_server['home'], path.strpath),
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    url, = urls_from_details(req.text)
    assert_url_matches_content(url, content)


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
@pytest.mark.usefixtures('cli_on_path')
def test_single_file_upload_from_stdin(content, running_server):
    info_url = subprocess.check_output(
        ('fput', '--server', running_server['home'], '-'),
        input=content,
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    url, = urls_from_details(req.text)
    assert_url_matches_content(url, content)


@pytest.mark.usefixtures('cli_on_path')
def test_multiple_file_upload(running_server, tmpdir):
    paths = []
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        path = tmpdir.join('ohai{}.bin'.format(i))
        path.write(content, 'wb')
        paths.append(path.strpath)

    info_url = subprocess.check_output(
        ('fput', '--server', running_server['home']) + tuple(paths),
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    urls = urls_from_details(req.text)
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        assert 'ohai{}.bin'.format(i) in req.text
        assert_url_matches_content(urls[i], content)
