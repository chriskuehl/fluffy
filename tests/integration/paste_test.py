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
