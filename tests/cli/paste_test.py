import subprocess

import pytest
import requests

from testing import assert_url_matches_content
from testing import PLAINTEXT_TESTCASES
from testing import raw_text_url_from_paste_html


@pytest.mark.parametrize('content', PLAINTEXT_TESTCASES)
@pytest.mark.usefixtures('cli_on_path')
def test_simple_paste_from_file(content, running_server, tmpdir):
    path = tmpdir.join('ohai.txt')
    path.write(content, 'w')
    info_url = subprocess.check_output(
        ('fpb', '--server', running_server['home'], path.strpath),
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    assert_url_matches_content(
        raw_text_url_from_paste_html(req.text),
        content.encode('UTF-8'),
    )


@pytest.mark.parametrize('content', PLAINTEXT_TESTCASES)
@pytest.mark.usefixtures('cli_on_path')
def test_simple_paste_from_stdin(content, running_server, tmpdir):
    info_url = subprocess.check_output(
        ('fpb', '--server', running_server['home']),
        input=content.encode('UTF-8'),
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    assert_url_matches_content(
        raw_text_url_from_paste_html(req.text),
        content.encode('UTF-8'),
    )


@pytest.mark.usefixtures('cli_on_path')
def test_paste_with_direct_link(running_server, tmpdir):
    info_url = subprocess.check_output(
        ('fpb', '--server', running_server['home'], '--direct-link'),
        input=b'hello world!',
    ).strip()

    req = requests.get(info_url)
    assert req.status_code == 200
    assert req.text == 'hello world!'


@pytest.mark.usefixtures('cli_on_path')
def test_paste_with_tee(running_server, tmpdir):
    input_text = b'hello\nworld!'
    output = subprocess.check_output(
        ('fpb', '--server', running_server['home'], '--tee'),
        input=input_text,
    ).strip()

    assert input_text in output
    info_url = output.split(b'\n')[-1]
    req = requests.get(info_url)
    assert req.status_code == 200

    assert_url_matches_content(
        raw_text_url_from_paste_html(req.text),
        input_text,
    )
