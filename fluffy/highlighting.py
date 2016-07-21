import pygments
from pygments.formatters import HtmlFormatter
from pygments.lexers import guess_lexer
from pygments.styles import get_style_by_name

from fluffy import app


_pygments_formatter = HtmlFormatter(
    noclasses=True,
    style=get_style_by_name('xcode'),
)


@app.template_filter()
def highlight(text):
    return pygments.highlight(
        text,
        guess_lexer(text),
        _pygments_formatter,
    )
