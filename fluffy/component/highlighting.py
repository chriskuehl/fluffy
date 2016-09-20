import re
from collections import namedtuple

import pygments
import pygments.lexers
from pygments.formatters import HtmlFormatter
from pygments.styles import get_style_by_name
from pyquery import PyQuery as pq


# We purposefully don't list all possible languages, and instead just the ones
# we think people are most likely to use.
UI_LANGUAGES_MAP = {
    'bash': 'Bash / Shell',
    'c': 'C',
    'c++': 'C++',
    'cheetah': 'Cheetah',
    'diff': 'Diff',
    'groovy': 'Groovy',
    'html': 'HTML',
    'java': 'Java',
    'javascript': 'JavaScript',
    'json': 'JSON',
    'makefile': 'Makefile',
    'objective-c': 'Objective-C',
    'php': 'PHP',
    'python3': 'Python',
    'ruby': 'Ruby',
    'scala': 'Scala',
    'sql': 'SQL',
    'yaml': 'YAML',
}


_pygments_formatter = HtmlFormatter(
    noclasses=True,
    linespans='line',
    nobackground=True,
    style=get_style_by_name('xcode'),
)


class PygmentsHighlighter(namedtuple('PygmentsHighlighter', ('lexer',))):

    @property
    def name(self):
        return self.lexer.name

    def highlight(self, text):
        return _highlight(text, self.lexer)


class DiffHighlighter(namedtuple('PygmentsHighlighter', ('lexer',))):

    @property
    def name(self):
        return 'Diff ({})'.format(self.lexer.name)

    def highlight(self, text):
        html = pq(_highlight(text, self.lexer))
        lines = html('pre > span')

        # there's an empty span at the start...
        assert 'id' not in lines[0].attrib
        pq(lines[0]).remove()
        lines.pop(0)

        for line in lines:
            line = pq(line)
            assert line.attr('id').startswith('line-')

            el = pq(line)

            # .text() doesn't include whitespace before it, but .html() does
            h = el.html()
            text = h[:len(h) - len(h.lstrip())] + el.text()

            if text.startswith('+'):
                line.addClass('diff-add')
            elif text.startswith('-'):
                line.addClass('diff-remove')

        return html.outerHtml()


def looks_like_diff(text):
    """Return whether the text looks like a diff."""
    # TODO: improve this
    return bool(re.search(r'^diff --git ', text, re.MULTILINE))


def strip_diff_things(text):
    """Remove things from the text that make it look like a diff.

    The purpose of this is so we can run guess_lexer over the source text. If
    we have a diff of Python, Pygments might tell us the language is "Diff".
    Really, we want it to highlight it like it's Python, and then we'll apply
    the diff formatting on top.
    """
    s = ''

    for line in text.splitlines():
        if line.startswith((
            'diff --git',
            '--- ',
            '+++ ',
            'index ',
            '@@ ',
            'Author:',
            'AuthorDate:',
            'Commit:',
            'CommitDate:',
            'commit ',
        )):
            continue

        if line.startswith(('+', '-')):
            line = line[1:]

        s += line + '\n'

    return s


def get_highlighter(text, language):
    lexer = guess_lexer(text, language)

    if (
            lexer is None or
            language is None or
            lexer.name.lower() != language.lower() or
            lexer.name.lower() == 'diff'
    ):
        # If it wasn't a perfect match, then we had to guess a language.
        # Pygments diff highlighting isn't too great, so we try to handle that
        # ourselves a bit.
        if lexer.name.lower() == 'diff' or looks_like_diff(text):
            return DiffHighlighter(
                guess_lexer(strip_diff_things(text), None),
            )

    return PygmentsHighlighter(lexer)


def guess_lexer(text, language, opts=None):
    lexer_opts = {'stripnl': False}
    if opts:
        lexer_opts = dict(lexer_opts, **opts)

    try:
        return pygments.lexers.get_lexer_by_name(language, **lexer_opts)
    except pygments.util.ClassNotFound:
        try:
            return pygments.lexers.guess_lexer(text, **lexer_opts)
        except pygments.util.ClassNotFound:
            return pygments.lexers.get_lexer_by_name('python', **lexer_opts)


def _highlight(text, lexer):
    return pygments.highlight(
        text,
        lexer,
        _pygments_formatter,
    )
