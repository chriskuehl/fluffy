import pygments.lexers
import pytest

from fluffy.component.highlighting import DiffHighlighter
from fluffy.component.highlighting import get_highlighter
from fluffy.component.highlighting import guess_lexer
from fluffy.component.highlighting import looks_like_diff
from fluffy.component.highlighting import PygmentsHighlighter
from fluffy.component.highlighting import strip_diff_things
from fluffy.component.highlighting import UI_LANGUAGES_MAP


EXAMPLE_DIFF = '''\
commit 5eb58ea2be01b451583429c4d8a931c0bcdbac8e
Author:     Chris Kuehl <ckuehl@ocf.berkeley.edu>
AuthorDate: Mon Jul 25 20:49:11 2016 -0400
Commit:     Chris Kuehl <ckuehl@ocf.berkeley.edu>
CommitDate: Mon Jul 25 20:49:11 2016 -0400

    Don't strip newlines, add horizontal scrollbar when overflow

diff --git a/fluffy/highlighting.py b/fluffy/highlighting.py
index 217363a..409d912 100644
--- a/fluffy/highlighting.py
+++ b/fluffy/highlighting.py
@@ -38,12 +38,12 @@ _pygments_formatter = HtmlFormatter(

 def guess_lexer(text, language):
     try:
-        return pygments.lexers.get_lexer_by_name(language)
+        return pygments.lexers.get_lexer_by_name(language, stripnl=False)
     except pygments.util.ClassNotFound:
         - try:
-            return pygments.lexers.guess_lexer(text)
+            return pygments.lexers.guess_lexer(text, stripnl=False)
         except pygments.util.ClassNotFound:
-            return pygments.lexers.get_lexer_by_name('python')
+            return pygments.lexers.get_lexer_by_name('python', stripnl=False)
'''


EXAMPLE_C = '''\
#include <stdio.h>
#include <stdlib.h>

int main(void);

int main(void) {
    uint8_t x = 42;
    uint8_t y = x + 1;

    /* exit 1 for success! */
    return 1;
}
'''


@pytest.mark.parametrize('language', UI_LANGUAGES_MAP)
def test_ui_language_exists(language):
    """Ensure a lexer exists for each language we advertise."""
    assert pygments.lexers.get_lexer_by_name('python') is not None


def test_guess_lexer_uses_valid_lang():
    assert guess_lexer(EXAMPLE_C, 'ruby').name == 'Ruby'


@pytest.mark.parametrize('invalid_lang', ['herpderp', '', None, 'autodetect'])
def test_guess_lexer_autodetects_with_invalid_lang(invalid_lang):
    assert guess_lexer(EXAMPLE_C, invalid_lang).name == 'C'


def test_guess_lexer_falls_back_to_python():
    assert guess_lexer('what language even is this', None).name == 'Python'


@pytest.mark.parametrize(('text', 'expected'), (
    ('', False),
    (
        'some simple\n'
        'text is here\n',
        False,
    ),
    (EXAMPLE_DIFF, True),
))
def test_looks_like_diff(text, expected):
    assert looks_like_diff(text) is expected


def test_strip_diff_things():
    assert strip_diff_things(EXAMPLE_DIFF) == '''\

    Don't strip newlines, add horizontal scrollbar when overflow


 def guess_lexer(text, language):
     try:
        return pygments.lexers.get_lexer_by_name(language)
        return pygments.lexers.get_lexer_by_name(language, stripnl=False)
     except pygments.util.ClassNotFound:
         - try:
            return pygments.lexers.guess_lexer(text)
            return pygments.lexers.guess_lexer(text, stripnl=False)
         except pygments.util.ClassNotFound:
            return pygments.lexers.get_lexer_by_name('python')
            return pygments.lexers.get_lexer_by_name('python', stripnl=False)
'''


@pytest.mark.parametrize(('text', 'language', 'expected'), (
    (EXAMPLE_C, 'c', pygments.lexers.get_lexer_by_name('c')),
    (EXAMPLE_C, 'does not exist', pygments.lexers.get_lexer_by_name('c')),
    (EXAMPLE_C, None, pygments.lexers.get_lexer_by_name('c')),
    (EXAMPLE_DIFF, 'c', pygments.lexers.get_lexer_by_name('c')),
))
def test_get_highlighter_pygments(text, language, expected):
    h = get_highlighter(text, language)
    assert isinstance(h, PygmentsHighlighter)
    assert type(h.lexer) is type(expected)


@pytest.mark.parametrize(('text', 'language', 'expected'), (
    (EXAMPLE_DIFF, None, pygments.lexers.get_lexer_by_name('python')),
    (EXAMPLE_DIFF, 'diff', pygments.lexers.get_lexer_by_name('python')),
    (EXAMPLE_C, 'diff', pygments.lexers.get_lexer_by_name('c')),
))
def test_get_highlighter_diff(text, language, expected):
    h = get_highlighter(text, language)
    assert isinstance(h, DiffHighlighter)
    assert type(h.lexer) is type(expected)
