import io
import re

import pytest
import requests


FILE_CONTENT_TESTCASES = (
    b'',
    b'hello world',
    b'hello\nworld\n',
    'éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ'.encode('utf8'),
    '\x43\x92\xd9\x0f\xaf\x32\x2c\x00\x12\x23'.encode('utf8'),
)


@pytest.mark.parametrize('content', FILE_CONTENT_TESTCASES)
def test_single_file_upload(content, running_server):
    req = requests.post(
        running_server['home'] + '/upload',
        files={'file': ('ohai.bin', io.BytesIO(content), None, None)},
    )
    assert req.status_code == 200
    assert 'ohai.bin' in req.text

    m = re.search(r'<a href="(http://localhost:\d+/object/[^"]+\.bin)"', req.text)
    assert m, req.text

    req = requests.get(m.group(1))
    assert req.content == content
