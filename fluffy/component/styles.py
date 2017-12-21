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

FG_COLORS_DARK = {
    'Black': '#555753',
    'Red': '#FF5C5C',
    'Green': '#8AE234',
    'Yellow': '#FCE94F',
    'Blue': '#8FB6E1',
    'Magenta': '#FF80F1',
    'Cyan': '#34E2E2',
    'White': '#EEEEEC',
}

BG_COLORS_DARK = {
    'Black': '#555753',
    'Red': '#F03D3D',
    'Green': '#6ABC1B',
    'Yellow': '#CEB917',
    'Blue': '#6392C6',
    'Magenta': '#FF80F1',
    'Cyan': '#2FC0C0',
    'White': '#BFBFBF',
}


def _make_style(
        *,
        name=None,
        base,
        fg_colors=FG_COLORS_LIGHT,
        bg_colors=BG_COLORS_LIGHT,
        toolbar_fg_color='#333',
        toolbar_bg_color='#e0e0e0',
        border_color='#eee',
        line_numbers_fg_color='#222',
        line_numbers_bg_color='#fafafa',
        line_numbers_hover_bg_color='#ffeaaf',
        line_numbers_selected_bg_color='#ffe18e',
        selected_line_bg_color='#fff3d3',
        diff_add_line_bg_color='#e2ffe2',
        diff_add_selected_line_bg_color='#e8ffbc',
        diff_remove_line_bg_color='#ffe5e5',
        diff_remove_selected_line_bg_color='#ffdfbf'
):
    base_style = pygments.styles.get_style_by_name(base)
    new_styles = dict(base_style.styles)
    new_styles.update(color_tokens(fg_colors, bg_colors))
    return type(
        'Fluffy' + base_style.__name__,
        (base_style,),
        {
            'name': name or base,
            'styles': new_styles,
            'toolbar_fg_color': toolbar_fg_color,
            'toolbar_bg_color': toolbar_bg_color,
            'border_color': border_color,
            'line_numbers_fg_color': line_numbers_fg_color,
            'line_numbers_bg_color': line_numbers_bg_color,
            'line_numbers_hover_bg_color': line_numbers_hover_bg_color,
            'line_numbers_selected_bg_color': line_numbers_selected_bg_color,
            'selected_line_bg_color': selected_line_bg_color,
            'diff_add_line_bg_color': diff_add_line_bg_color,
            'diff_add_selected_line_bg_color': diff_add_selected_line_bg_color,
            'diff_remove_line_bg_color': diff_remove_line_bg_color,
            'diff_remove_selected_line_bg_color': diff_remove_selected_line_bg_color,
        },
    )


DEFAULT_STYLE = _make_style(
    name='default',
    base='xcode',
)

STYLES_BY_CATEGORY = {
    'Light': sorted(
        (
            DEFAULT_STYLE,
            _make_style(
                base='pastie',
            ),
        ),
        key=operator.attrgetter('name'),
    ),
    'Dark': sorted(
        (
            _make_style(
                base='monokai',
                fg_colors=FG_COLORS_DARK,
                bg_colors=BG_COLORS_DARK,
                border_color='#454545',
                line_numbers_fg_color='#999',
                line_numbers_bg_color='#272822',
                line_numbers_hover_bg_color='#8D8D8D',
                line_numbers_selected_bg_color='#5F5F5F',
                selected_line_bg_color='#545454',
                diff_add_line_bg_color='#3d523d',
                diff_add_selected_line_bg_color='#607b60',
                diff_remove_line_bg_color='#632727',
                diff_remove_selected_line_bg_color='#9e4848',
            ),
            _make_style(
                base='solarizeddark',
                fg_colors=FG_COLORS_DARK,
                bg_colors=BG_COLORS_DARK,
                border_color='#454545',
                line_numbers_fg_color='#656565',
                line_numbers_bg_color='#002b36',
                line_numbers_hover_bg_color='#00596f',
                line_numbers_selected_bg_color='#004252',
                selected_line_bg_color='#004252',
                diff_add_line_bg_color='#0e400e',
                diff_add_selected_line_bg_color='#176117',
                diff_remove_line_bg_color='#632727',
                diff_remove_selected_line_bg_color='#9e4848',
            ),
        ),
        key=operator.attrgetter('name'),
    ),
}


def main():
    for styles in STYLES_BY_CATEGORY.values():
        for style in styles:
            prefix = '.highlight-{}'.format(style.name)
            print(HtmlFormatter(style=style).get_style_defs(prefix + ' .highlight'))
            print(
                '{prefix} .line-numbers {{'
                '  background-color: {style.line_numbers_bg_color};'
                '  border-color: {style.border_color};'
                '}}'
                '{prefix} .text {{'
                '  background-color: {style.border_color};'
                '}}'
                '{prefix} .line-numbers a {{'
                '  color: {style.line_numbers_fg_color};'
                '}}'
                '{prefix} .line-numbers a:hover {{'
                '  background-color: {style.line_numbers_hover_bg_color} !important;'
                '}}'
                '{prefix} .line-numbers a.selected {{'
                '  background-color: {style.line_numbers_selected_bg_color};'
                '}}'
                '{prefix} .paste-toolbar {{'
                '  background-color: {style.toolbar_bg_color};'
                '  color: {style.toolbar_fg_color};'
                '}}'
                '{prefix} .text .highlight > pre > span.selected {{'
                '  background-color: {style.selected_line_bg_color};'
                '}}'
                '{prefix} .text .highlight > pre > span.diff-add {{'
                '  background-color: {style.diff_add_line_bg_color};'
                '}}'
                '{prefix} .text .highlight > pre > span.diff-add.selected {{'
                '  background-color: {style.diff_add_selected_line_bg_color};'
                '}}'
                '{prefix} .text .highlight > pre > span.diff-remove {{'
                '  background-color: {style.diff_remove_line_bg_color};'
                '}}'
                '{prefix} .text .highlight > pre > span.diff-remove.selected {{'
                '  background-color: {style.diff_remove_selected_line_bg_color};'
                '}}'
                ''.format(prefix=prefix, style=style),
            )


if __name__ == '__main__':
    exit(main())
