import re
from collections import namedtuple

import pygments.lexers
import pygments.styles.xcode
from pygments.formatters import HtmlFormatter
from pyquery import PyQuery as pq

from fluffy.component.styles import DEFAULT_STYLE


# We purposefully don't list all possible languages, and instead just the ones
# we think people are most likely to use.
UI_LANGUAGES_MAP = {
    'bash': 'Bash / Shell',
    'c': 'C',
    'c++': 'C++',
    'cheetah': 'Cheetah',
    'diff': 'Diff',
    'groovy': 'Groovy',
    'haskell': 'Haskell',
    'html': 'HTML',
    'java': 'Java',
    'javascript': 'JavaScript',
    'json': 'JSON',
    'kotlin': 'Kotlin',
    'lua': 'Lua',
    'makefile': 'Makefile',
    'objective-c': 'Objective-C',
    'php': 'PHP',
    'python3': 'Python',
    'ruby': 'Ruby',
    'rust': 'Rust',
    'scala': 'Scala',
    'sql': 'SQL',
    'swift': 'Swift',
    'yaml': 'YAML',
}


_pygments_formatter = HtmlFormatter(
    noclasses=False,
    linespans='line',
    style=DEFAULT_STYLE,
)


class PygmentsHighlighter(namedtuple('PygmentsHighlighter', ('lexer',))):

    @property
    def name(self):
        return self.lexer.name

    def highlight(self, text):
        text = _highlight(text, self.lexer)
        return text


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


def looks_like_ansi_color(text):
    """Return whether the text looks like it has ANSI color codes."""
    return '\x1b[' in text


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


def get_highlighter(text, language, filename):
    if language in {None, 'autodetect'} and looks_like_ansi_color(text):
        language = 'ansi-color'

    lexer = guess_lexer(text, language, filename)

    diff_requested = (language or '').startswith('diff-')

    if (
            diff_requested or
            lexer is None or
            language is None or
            lexer.name.lower() != language.lower() or
            lexer.name.lower() == 'diff'
    ):
        if diff_requested:
            _, requested_diff_language = language.split('-', 1)
        else:
            requested_diff_language = None

        # If it wasn't a perfect match, then we had to guess a language.
        # Pygments diff highlighting isn't too great, so we try to handle that
        # ourselves a bit.
        if diff_requested or lexer.name.lower() == 'diff' or looks_like_diff(text):
            return DiffHighlighter(
                guess_lexer(strip_diff_things(text), requested_diff_language, filename),
            )

    return PygmentsHighlighter(lexer)


def guess_lexer(text, language, filename, opts=None):
    lexer_opts = {'stripnl': False}
    if opts:
        lexer_opts = dict(lexer_opts, **opts)

    # First, look for an exact lexer match name.
    try:
        return pygments.lexers.get_lexer_by_name(language, **lexer_opts)
    except pygments.util.ClassNotFound:
        pass

    # If that didn't work, if given a file name, try finding a lexer using that.
    if filename is not None:
        try:
            return pygments.lexers.guess_lexer_for_filename(filename, text, **lexer_opts)
        except pygments.util.ClassNotFound:
            pass

    # Finally, try to guess by looking at the file content.
    try:
        lexer = pygments.lexers.guess_lexer(text, **lexer_opts)

        # Newer versions of Pygments will virtually always fall back to
        # TextLexer due to its 0.01 priority (which is what it returns on
        # analyzing any text).
        if not isinstance(lexer, pygments.lexers.TextLexer):
            return lexer
    except pygments.util.ClassNotFound:
        pass

    # Default to Python, it highlights most things reasonably.
    return pygments.lexers.get_lexer_by_name('python', **lexer_opts)


def _highlight(text, lexer):
    return pygments.highlight(
        text,
        lexer,
        _pygments_formatter,
    )
