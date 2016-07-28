import subprocess

import pytest
import requests

from testing import assert_url_matches_content
from testing import FILE_CONTENT_TESTCASES
from testing import urls_from_details


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
def test_single_file_upload(content, running_server, tmpdir):
    path = tmpdir.join('ohai.bin')
    path.write(content)
    info_url = subprocess.check_output(
        ('fput', '--server', running_server['home'], path.strpath),
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    url, = urls_from_details(req.text)
    assert_url_matches_content(url, content)


def test_multiple_file_upload(running_server, tmpdir):
    paths = []
    for i, content in enumerate(FILE_CONTENT_TESTCASES):
        path = tmpdir.join('ohai{}.bin'.format(i))
        path.write(content)
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
