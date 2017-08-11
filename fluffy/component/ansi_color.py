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


def callback(lexer, match):
    codes = match.group(1).split(';')

    for code in codes:
        try:
            code = int(code)
        except ValueError:
            # This shouldn't ever happen if we're given valid ANSI color codes,
            # but people can pastebin junk, and we should tolerate that.
            continue
        else:
            fg_color = _ansi_code_to_color.get(code - 30)
            bg_color = _ansi_code_to_color.get(code - 40)
            if fg_color:
                lexer.fg_color = fg_color
            elif bg_color:
                lexer.bg_color = bg_color
            elif code == 1:
                lexer.bold = True
            elif code == 22:
                lexer.bold = False
            elif code == 39:
                lexer.fg_color = None
            elif code == 49:
                lexer.bg_color = None
            elif code == 0:
                lexer.reset_state()

    yield match.start(), lexer.current_token(), match.group(2)


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
            (r'\x1b\[(.*?)m([^\x1b]*)', callback),
        ],
    }
