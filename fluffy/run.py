from pathlib import Path

from flask import render_template

from fluffy.app import app
from fluffy.component.highlighting import get_highlighter
from fluffy.component.styles import STYLES_BY_CATEGORY
from fluffy.models import HtmlToStore
from fluffy.models import UploadedFile

# imports so the decorators run :(
import fluffy.component.markdown  # noreorder # noqa
import fluffy.views  # noreorder # noqa


TESTING_DIR = Path(__file__).parent.parent / 'testing'


class UploadedFileWithFakeURL(UploadedFile):

    def __new__(cls, **kwargs):
        url = kwargs.pop('url')
        created = super().__new__(cls, **kwargs)
        created.url = url
        return created


def debug():  # pragma: no cover

    @app.route('/test/paste')
    def view_paste():
        text = (TESTING_DIR / 'files' / 'code.py').open().read()
        highlighter = get_highlighter('', 'python', None)
        return render_template(
            'paste.html',
            texts=highlighter.prepare_text(text),
            highlighter=highlighter,
            edit_url='#edit',
            raw_url='#raw',
            styles=STYLES_BY_CATEGORY,
        )

    @app.route('/test/diff')
    def view_diff():
        text = (TESTING_DIR / 'files' / 'python.diff').open().read()
        highlighter = get_highlighter(text, None, None)
        return render_template(
            'paste.html',
            texts=highlighter.prepare_text(text),
            highlighter=highlighter,
            edit_url='#edit',
            raw_url='#raw',
            styles=STYLES_BY_CATEGORY,
        )

    @app.route('/test/ansi-color')
    def view_ansi_color():
        text = (TESTING_DIR / 'files' / 'ansi-color').open().read()
        highlighter = get_highlighter(text, None, None)
        return render_template(
            'paste.html',
            texts=highlighter.prepare_text(text),
            highlighter=highlighter,
            edit_url='#edit',
            raw_url='#raw',
            styles=STYLES_BY_CATEGORY,
        )

    @app.route('/test/markdown')
    def view_markdown():
        return render_template(
            'markdown.html',
            text=(TESTING_DIR / 'files' / 'markdown.md').open().read(),
            edit_url='#edit',
            raw_url='#raw',
        )

    @app.route('/test/upload-details')
    def view_upload_details():
        return render_template(
            'details.html',
            uploads=[
                # Pictures
                (
                    UploadedFileWithFakeURL(
                        human_name='Screenshot_2022-05-09_15-01-12.png',
                        num_bytes=1234567,
                        open_file=None,
                        unique_id='asdf',
                        url='https://i.fluffy.cc/wkwtqwkjvFmj4pxSfwj6W126Bb8ggQnw.jpeg',
                    ),
                    None,
                ),
                (
                    UploadedFileWithFakeURL(
                        human_name='another picture.png',
                        num_bytes=1234567,
                        open_file=None,
                        unique_id='asdf',
                        url='https://i.fluffy.cc/RMQvJSHdC7JXh4nWXH5VWQpGkqhfTDk7.png',
                    ),
                    None,
                ),
                (
                    UploadedFileWithFakeURL(
                        human_name='amog.png',
                        num_bytes=1234567,
                        open_file=None,
                        unique_id='asdf',
                        url='https://i.fluffy.cc/3r9PtfR7C5qMf4p1Q9dSsHdsBdXKRQMg.png',
                    ),
                    None,
                ),
                # Text with paste view
                (
                    UploadedFile(
                        human_name='foo.txt',
                        num_bytes=1234,
                        open_file=None,
                        unique_id='UNIQUE_ID',
                    ),
                    HtmlToStore(
                        name='asdf',
                        open_file=None,
                    ),
                ),
                # Non-text non-displayable extension
                (
                    UploadedFile(
                        human_name='my cool document.xls',
                        num_bytes=44234343,
                        open_file=None,
                        unique_id='UNIQUE_ID',
                    ),
                    None,
                ),
                # Non-text very long file name
                (
                    UploadedFile(
                        human_name='my cool document asdfasdf fasdfsdf jsdfasjfasdf asdfasdf asdfsdfd.xls',
                        num_bytes=44234343,
                        open_file=None,
                        unique_id='UNIQUE_ID',
                    ),
                    None,
                ),
            ],
        )

    app.run(debug=True)


if __name__ == '__main__':
    exit(debug())
