import operator

import pygments.styles
from pygments.formatters import HtmlFormatter
from pygments_ansi_color import color_tokens


FG_COLORS_LIGHT = {
    'Black': '#000000',
    'Red': '#EF2929',
    'Green': '#62ca00',
    'Yellow': '#dac200',
    'Blue': '#3465A4',
    'Magenta': '#ce42be',
    'Cyan': '#34E2E2',
    'White': '#ffffff',
}

BG_COLORS_LIGHT = {
    'Black': '#000000',
    'Red': '#EF2929',
    'Green': '#8AE234',
    'Yellow': '#FCE94F',
    'Blue': '#3465A4',
    'Magenta': '#c509c5',
    'Cyan': '#34E2E2',
    'White': '#ffffff',
}


def _make_style(*, name=None, base, fg_colors, bg_colors):
    base_style = pygments.styles.get_style_by_name(base)
    new_styles = dict(base_style.styles)
    new_styles.update(color_tokens(fg_colors, bg_colors))
    return type(
        'Fluffy' + base_style.__name__,
        (base_style,),
        {
            'name': name or base,
            'styles': new_styles,
        },
    )


DEFAULT_STYLE = _make_style(
    name='default',
    base='xcode',
    fg_colors=FG_COLORS_LIGHT,
    bg_colors=BG_COLORS_LIGHT,
)

STYLES_BY_CATEGORY = {
    'Light': sorted(
        (
            DEFAULT_STYLE,
            _make_style(
                base='pastie',
                fg_colors=FG_COLORS_LIGHT,
                bg_colors=BG_COLORS_LIGHT,
            ),
        ),
        key=operator.attrgetter('name'),
    ),
    'Dark': sorted(
        (
            _make_style(
                base='monokai',
                fg_colors=FG_COLORS_LIGHT,
                bg_colors=BG_COLORS_LIGHT,
            ),
            _make_style(
                base='native',
                fg_colors=FG_COLORS_LIGHT,
                bg_colors=BG_COLORS_LIGHT,
            ),
        ),
        key=operator.attrgetter('name'),
    ),
}


def main():
    for styles in STYLES_BY_CATEGORY.values():
        for style in styles:
            print(HtmlFormatter(style=style).get_style_defs('.highlight-{} .highlight'.format(style.name)))


if __name__ == '__main__':
    exit(main())
