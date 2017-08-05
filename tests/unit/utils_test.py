import pytest

from fluffy.utils import content_is_binary
from fluffy.utils import gen_unique_id
from fluffy.utils import human_size
from fluffy.utils import pluralize


@pytest.mark.parametrize(
    ('num', 'expected'), [
        (-4, 'things'),
        (-1, 'thing'),
        (0, 'things'),
        (1, 'thing'),
        (4, 'things'),
    ],
)
def test_pluralize(num, expected):
    assert pluralize('thing', num)


@pytest.mark.parametrize(
    ('num_bytes', 'expected'), [
        (0, '0 bytes'),
        (1, '1 byte'),
        (42, '42 bytes'),
        (1500, '1.5 KiB'),
        (1500, '1.5 KiB'),
        (43123150, '41.1 MiB'),
        (8123000222, '7.6 GiB'),
    ],
)
def test_human_size(num_bytes, expected):
    assert human_size(num_bytes) == expected


def test_gen_unique_uid():
    id1 = gen_unique_id()
    assert len(id1) == 32

    id2 = gen_unique_id()
    assert len(id2) == 32

    assert id1 != id2


@pytest.mark.parametrize(
    ('data', 'expected'), [
        (b'hello world', False),
        (b'hello\nworld\nherp\nderp', False),
        (b'', False),
        ('éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ'.encode('utf8'), False),
        ('¯\_(ツ)_/¯'.encode('utf8'), False),
        ('♪┏(・o･)┛♪┗ ( ･o･) ┓♪┏ ( ) ┛♪┗ (･o･ ) ┓♪┏(･o･)┛♪'.encode('utf8'), False),
        ('éóñå'.encode('latin1'), False),

        (b'hello world\x00', True),
        (b'\x7f\x45\x4c\x46\x02\x01\x01', True),  # first few bytes of /bin/bash
        (b'\x43\x92\xd9\x0f\xaf\x32\x2c', True),  # some /dev/urandom output
    ],
)
def test_content_is_binary(data, expected):
    assert content_is_binary(data) == expected
