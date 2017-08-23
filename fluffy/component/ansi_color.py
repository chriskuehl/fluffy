import re

import pygments.lexer
import pygments.token
from pygments.token import Text


# Pygments doesn't have a generic "color" token; instead everything is
# contextual (e.g. "comment" or "variable"). That doesn't make sense for us,
# where the actual colors actually *are* what we care about.
Color = pygments.token.Token.Color

_ansi_code_to_color = {
    0: 'Black',
    1: 'Red',
    2: 'Green',
    3: 'Yellow',
    4: 'Blue',
    5: 'Magenta',
    6: 'Cyan',
    7: 'White',
}


def token_from_lexer_state(bold, fg_color, bg_color):
    """Construct a token given the current lexer state.

    We can only emit one token even though we have a multiple-tuple state.

    To do this, we construct tokens like "BoldRed".
    """
    token_name = ''

    if bold:
        token_name += 'Bold'

    if fg_color:
        token_name += fg_color

    if bg_color:
        token_name += 'BG' + bg_color

    if token_name == '':
        return Text
    else:
        return getattr(Color, token_name)


def _callback(lexer, match):
    # http://ascii-table.com/ansi-escape-sequences.php
    # "after_escape" contains everything after the escape sequence, up to the
    # next escape sequence. We still need to separate the content from the end
    # of the escape sequence.
    after_escape = match.group(1)

    # TODO: this doesn't handle the case where the values are non-numeric. This
    # is rare but can happen for keyboard remapping, e.g. '\x1b[0;59;"A"p'
    parsed = re.match(r'([0-9;=]*?)?([a-zA-Z])(.*)$', after_escape, re.DOTALL | re.MULTILINE)
    if parsed is None:
        # This shouldn't ever happen if we're given valid ANSI color codes,
        # but people can pastebin junk, and we should tolerate that.
        text = after_escape
    else:
        value, code, text = parsed.group(1), parsed.group(2), parsed.group(3)

        # We currently only handle "m" ("Set Graphics Mode")
        if code == 'm':
            values = value.split(';')
            for value in values:
                try:
                    value = int(value)
                except ValueError:
                    # Shouldn't ever happen, but could with invalid ANSI codes.
                    continue
                else:
                    fg_color = _ansi_code_to_color.get(value - 30)
                    bg_color = _ansi_code_to_color.get(value - 40)
                    if fg_color:
                        lexer.fg_color = fg_color
                    elif bg_color:
                        lexer.bg_color = bg_color
                    elif value == 1:
                        lexer.bold = True
                    elif value == 22:
                        lexer.bold = False
                    elif value == 39:
                        lexer.fg_color = None
                    elif value == 49:
                        lexer.bg_color = None
                    elif value == 0:
                        lexer.reset_state()

    yield match.start(), lexer.current_token(), text


class AnsiColorLexer(pygments.lexer.RegexLexer):
    name = 'ANSI Terminal'
    flags = re.DOTALL | re.MULTILINE

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.reset_state()

    def current_token(self):
        return token_from_lexer_state(
            self.bold, self.fg_color, self.bg_color,
        )

    def reset_state(self):
        self.bold = False
        self.fg_color = None
        self.bg_color = None

    tokens = {
        'root': [
            (r'\x1b\[([^\x1b]*)', _callback),
        ],
    }
