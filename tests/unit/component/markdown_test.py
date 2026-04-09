import pytest

from fluffy.component.highlighting import looks_like_markdown


REAL_MARKDOWN = '''\
# My Document

This is a paragraph with a [link](https://example.com).

## Section Two

```python
print("hello")
```
'''

FRONTMATTER_ONLY = '''\
---
title: My Post
date: 2024-01-01
---

Just some text here.
'''

SINGLE_HEADER = '''\
# Title

Just a title and some plain text below it.
'''

PLAIN_TEXT = '''\
The quick brown fox jumps over the lazy dog.
Nothing special about this text at all.
'''

LOG_OUTPUT = '''\
2024-01-01 12:00:00 INFO Starting server on port 8080
2024-01-01 12:00:01 WARN * connection timeout after 30s *
2024-01-01 12:00:02 ERROR ** fatal: out of memory **
'''

BOLD_AND_HEADER = '''\
# Configuration Guide

Set **debug** to true and **verbose** to false.
'''


@pytest.mark.parametrize(
    ('text', 'expected'),
    (
        (REAL_MARKDOWN, True),
        (FRONTMATTER_ONLY, True),
        (BOLD_AND_HEADER, True),
        (SINGLE_HEADER, False),
        (PLAIN_TEXT, False),
        (LOG_OUTPUT, False),
        ('', False),
    ),
)
def test_looks_like_markdown(text, expected):
    assert looks_like_markdown(text) is expected


def test_looks_like_markdown_only_checks_head():
    padding = 'x' * 10000
    late_markdown = padding + '# Header\n\n[link](url)\n'
    assert looks_like_markdown(late_markdown) is False
