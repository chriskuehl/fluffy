import collections
import dataclasses
import re
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


@dataclasses.dataclass(frozen=True)
class PasteText:
    text: str
    line_number_mapping: dict[int, int] | None = None


class PygmentsHighlighter(namedtuple('PygmentsHighlighter', ('lexer',))):

    is_diff = False

    @property
    def name(self):
        return self.lexer.name

    @property
    def is_terminal_output(self):
        return 'ansi-color' in self.lexer.aliases

    def prepare_text(self, text: str) -> list[PasteText]:
        return [PasteText(text)]

    def highlight(self, text, line_number_mapping: dict[int, int] | None = None):
        text = _highlight(text, self.lexer, line_number_mapping)
        return text


class DiffHighlighter(namedtuple('DiffHighlighter', ('lexer',))):

    is_terminal_output = False
    is_diff = True

    @property
    def name(self):
        return f'Diff ({self.lexer.name})'

    def prepare_text(self, text: str) -> list[PasteText]:
        """Transform the unified diff into a side-by-side diff."""
        diff1 = []
        diff2 = []

        line_number_mapping: dict[int, list[int]] = collections.defaultdict(list)

        def _fill_empty_lines():
            """Fill either side of the diff with empty lines so they have the same length."""
            diff1.extend([''] * max(0, (len(diff2) - len(diff1))))
            diff2.extend([''] * max(0, (len(diff1) - len(diff2))))

        for i, line in enumerate(text.splitlines()):
            if line.startswith('-'):
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

            line_number_mapping[max(len(diff1), len(diff2))].append(i + 1)

        _fill_empty_lines()
        assert len(diff1) == len(diff2), (len(diff1), len(diff2))
        return [
            PasteText('\n'.join(diff1), line_number_mapping),
            PasteText('\n'.join(diff2), line_number_mapping),
            PasteText(text),
        ]

    def highlight(self, text, line_number_mapping: dict[int, int] | None = None):
        html = pq(_highlight(text, self.lexer, line_number_mapping))
        lines = html('pre > span')

        # there's an empty span at the start...
        assert 'id' not in lines[0].attrib
        pq(lines[0]).remove()
        lines.pop(0)

        for line in lines:
            line = pq(line)
            assert line.attr('class').startswith('line-')

            el = pq(line)

            # This is not a perfect reconstruction of the text (e.g. doesn't
            # handle HTML entities), but is enough to get the first character
            # without running into
            # https://github.com/chriskuehl/fluffy/issues/150
            text = re.sub(r'<.*?>', '', el.html())
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
        if not (
            isinstance(lexer, pygments.lexers.TextLexer) or
            # Seems to flag for everything in recent Pygments...
            isinstance(lexer, pygments.lexers.ScdocLexer)
        ):
            return lexer
    except pygments.util.ClassNotFound:
        pass

    # Default to Python, it highlights most things reasonably.
    return pygments.lexers.get_lexer_by_name('python', **lexer_opts)


def _highlight(text, lexer, line_number_mapping: dict[int, int] | None):
    text = pygments.highlight(
        # Pygments works most consistently when the text is a series of lines
        # ending with newlines. The rest of fluffy is not treating lines that
        # way (instead, \n means start a new line which should actually be
        # displayed) so we always append a newline here.
        #
        # This has no effect if the final line of the text has characters on
        # it, but if the last line is empty (i.e. final character is already a
        # newline in `text`), Pygments will chop off the last line unless we do
        # this.
        text + '\n' if len(text) > 0 else text,
        lexer,
        _pygments_formatter,
    )

    def rewrite_line_id(match: re.Match[str]) -> str:
        line = int(match.group(1))
        if line_number_mapping is not None:
            lines = line_number_mapping[line]
        else:
            lines = [line]

        return 'class="{}"'.format(
            ' '.join(
                f'line-{line}' for line in lines
            ),
        )

    # Transform line IDs to classes and (if necessary) map line numbers.
    text = re.sub(r'id="line-(\d+)"', rewrite_line_id, text)

    return text
