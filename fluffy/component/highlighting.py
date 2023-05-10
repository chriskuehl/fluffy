import functools
import re
import typing
from collections import namedtuple

import pygments.lexers.teraterm
import pygments.styles.xcode
from pygments.formatters import HtmlFormatter
from pygments_ansi_color import ExtendedColorHtmlFormatterMixin
from pyquery import PyQuery as pq

from fluffy.component.styles import DEFAULT_STYLE
# Work around https://github.com/chriskuehl/fluffy/issues/88.
pygments.lexers.teraterm.TeraTermLexer.analyse_text = lambda _: -100


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


class FluffyFormatter(ExtendedColorHtmlFormatterMixin, HtmlFormatter):
    pass


_pygments_formatter = FluffyFormatter(
    noclasses=False,
    linespans='line',
    style=DEFAULT_STYLE,
)


class PygmentsHighlighter(namedtuple('PygmentsHighlighter', ('lexer',))):

    is_diff = False

    @property
    def name(self):
        return self.lexer.name

    @property
    def is_terminal_output(self):
        return 'ansi-color' in self.lexer.aliases

    def prepare_text(self, text: str) -> typing.List[str]:
        return [text]

    def highlight(self, text):
        text = _highlight(text, self.lexer)
        return text


class DiffHighlighter(namedtuple('DiffHighlighter', ('lexer',))):

    is_terminal_output = False
    is_diff = True

    @property
    def name(self):
        return f'Diff ({self.lexer.name})'

    def prepare_text(self, text: str) -> typing.List[str]:
        """Transform the unified diff into a side-by-side diff."""
        diff1 = []
        diff2 = []

        def _fill_empty_lines():
            """Fill either side of the diff with empty lines so they have the same length."""
            diff1.extend([''] * max(0, (len(diff2) - len(diff1))))
            diff2.extend([''] * max(0, (len(diff1) - len(diff2))))

        for line in text.splitlines():
            if line in ('--- ', '+++ '):
                pass
            elif line.startswith('-'):
                diff1.append(line)
            elif line.startswith('+'):
                diff2.append(line)
            else:
                # Fill empty lines so that both sides are "caught up" after any
                # additions/deletions and we can print the lines both sides
                # contain side-by-side.
                _fill_empty_lines()

                diff1.append(line)
                diff2.append(line)

        _fill_empty_lines()
        assert len(diff1) == len(diff2), (len(diff1), len(diff2))
        return ['\n'.join(diff1), '\n'.join(diff2), text]

    def highlight(self, text):
        html = pq(_highlight(text, self.lexer))
        lines = html('pre > span')

        # there's an empty span at the start...
        assert 'id' not in lines[0].attrib
        pq(lines[0]).remove()
        lines.pop(0)

        for line in lines:
            line = pq(line)
            assert line.attr('class').startswith('line-')

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


# All guesslang titles with available Pygments lexers match automatically
# except for these exceptions.
GUESSLANG_LANGUAGE_MAP = {
    'Batchfile': 'batch',
    'Visual Basic': 'vbscript',
}


@functools.lru_cache()
def _guesslang_guesser():
    try:
        # This is expensive (even just the import) so we do it at runtime
        # rather than import time.
        import guesslang
        return guesslang.Guess()
    except ImportError:
        return None


def _guesslang_guess_lexer(text, lexer_opts):
    guess = _guesslang_guesser()
    if guess is not None:
        # guesslang by default ensures that "The predicted language probability
        # must be higher than 2 standard deviations from the mean." to return a
        # guess, but in my testing this is pretty generous and almost
        # everything gets detected as _something_ (e.g. INI) which messes up
        # the highlighting. Adding an arbitrary requirement seems to help.
        probabilities = guess.probabilities(text)
        lang_title, probability = probabilities[0]
        print(f'Guessed {lang_title} @ {probability * 100:.2f}%')
        if probability < 0.3:
            return None
        if lang_title is not None:
            lang = GUESSLANG_LANGUAGE_MAP.get(lang_title, lang_title)
            try:
                return pygments.lexers.get_lexer_by_name(lang, **lexer_opts)
            except pygments.util.ClassNotFound:
                pass


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
        # guesslang (if available) does better than pygments so we try it first.
        lexer = _guesslang_guess_lexer(text, lexer_opts) or pygments.lexers.guess_lexer(text, **lexer_opts)

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
    text = pygments.highlight(
        text,
        lexer,
        _pygments_formatter,
    )
    # We may have multiple renders per page, but for some reason
    # Pygments's HtmlFormatter only supports IDs and not classes for
    # line numbers.
    text = text.replace(' id="line-', ' class="line-')
    return text
