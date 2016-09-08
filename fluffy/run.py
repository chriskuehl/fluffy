from pathlib import Path

from flask import render_template

from fluffy.app import app
from fluffy.component.highlighting import get_highlighter

# imports so the decorators run :(
import fluffy.component.markdown  # noreorder # noqa
import fluffy.views  # noreorder # noqa


TESTING_DIR = Path(__file__).parent.parent / 'testing'


def debug():  # pragma: no cover

    @app.route('/test/paste')
    def view_paste():
        return render_template(
            'paste.html',
            text=(TESTING_DIR / 'files' / 'code.py').open().read(),
            highlighter=get_highlighter('', 'python'),
            edit_url='#edit',
            raw_url='#raw',
        )

    @app.route('/test/diff')
    def view_diff():
        text = (TESTING_DIR / 'files' / 'python.diff').open().read()
        return render_template(
            'paste.html',
            text=text,
            highlighter=get_highlighter(text, None),
            edit_url='#edit',
            raw_url='#raw',
        )

    @app.route('/test/markdown')
    def view_markdown():
        return render_template(
            'markdown.html',
            text=(TESTING_DIR / 'files' / 'markdown.md').open().read(),
            edit_url='#edit',
            raw_url='#raw',
        )

    app.run(debug=True)


if __name__ == '__main__':
    exit(debug())
