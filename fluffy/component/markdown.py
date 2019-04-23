import re

import mistune

from fluffy.app import app
from fluffy.component.highlighting import guess_lexer
from fluffy.component.highlighting import PygmentsHighlighter


class HtmlCommentsInlineLexerMixin:
    """Strip HTML comments inside lines."""

    def enable_html_comments(self):
        self.rules.html_comment = re.compile(
            '^<!--(.*?)-->',
        )
        self.default_rules.insert(0, 'html_comment')

    def output_html_comment(self, m):
        return ''


class HtmlCommentsBlockLexerMixin:
    """Strip blocks which consist entirely of HTML comments."""

    def enable_html_comments(self):
        self.rules.html_comment = re.compile(
            '^<!--(.*?)-->',
        )
        self.default_rules.insert(0, 'html_comment')

    def parse_html_comment(self, m):
        pass


class CodeRendererMixin:
    """Render highlighted code."""

    def block_code(self, code, lang):
        return PygmentsHighlighter(
            guess_lexer(code, lang, None, opts={'stripnl': True}),
        ).highlight(code)


class FluffyMarkdownRenderer(
    CodeRendererMixin,
    mistune.Renderer,
):
    pass


class FluffyMarkdownInlineLexer(
    mistune.InlineLexer,
    HtmlCommentsInlineLexerMixin,
):
    pass


class FluffyMarkdownBlockLexer(
    mistune.BlockLexer,
    HtmlCommentsBlockLexerMixin,
):
    pass


_renderer = FluffyMarkdownRenderer(
    escape=True,
    hard_wrap=False,
)

_inline = FluffyMarkdownInlineLexer(_renderer)
_inline.enable_html_comments()

_block = FluffyMarkdownBlockLexer(mistune.BlockGrammar())
_block.enable_html_comments()


@app.template_filter()
def markdown(text):
    return mistune.Markdown(
        renderer=_renderer,
        inline=_inline,
        block=_block,
    )(text)
