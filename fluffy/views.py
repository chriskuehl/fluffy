import contextlib

from flask import jsonify
from flask import redirect
from flask import render_template
from flask import request

from fluffy.app import app
from fluffy.component.backends import get_backend
from fluffy.component.highlighting import get_highlighter
from fluffy.component.highlighting import UI_LANGUAGES_MAP
from fluffy.models import FileTooLargeError
from fluffy.models import HtmlToStore
from fluffy.models import UploadedFile
from fluffy.utils import human_size
from fluffy.utils import ICON_EXTENSIONS
from fluffy.utils import ONE_MB


@app.route('/', methods={'GET', 'POST'})
def home():
    text = request.form.get('text', '') or request.args.get('text', '')
    return render_template(
        'home.html',
        languages=sorted(
            UI_LANGUAGES_MAP.items(),
            key=lambda key_val: key_val[1],
        ),
        text=text,
        extra_html_classes='start-on-paste' if (text or 'text' in request.args) else '',
        icon_extensions=ICON_EXTENSIONS,
        max_upload_size=app.config['MAX_UPLOAD_SIZE'],
    )


def upload_objects(objects):
    links = sorted(obj.url for obj in objects)
    for obj in objects:
        if isinstance(obj, HtmlToStore):
            get_backend().store_html(obj, links)
        else:
            get_backend().store_object(obj, links)


@app.route('/upload', methods={'POST'})
def upload():
    """Process an upload and return JSON status."""
    uploaded_files = []

    with contextlib.ExitStack() as ctx:
        objects = []

        for f in request.files.getlist('file'):
            try:
                uf = ctx.enter_context(UploadedFile.from_http_file(f))
                objects.append(uf)

                # If it looks like text, make a pastebin as well.
                pb = None
                if uf.num_bytes < ONE_MB and not uf.probably_binary:
                    # We can't know for sure it's utf8, so this might fail.
                    # If so, we just won't make a pastebin for this file.
                    try:
                        text = uf.full_content.decode('utf8')
                    except UnicodeDecodeError:
                        pass
                    else:
                        pb = ctx.enter_context(HtmlToStore.from_html(render_template(
                            'paste.html',
                            text=text,
                            highlighter=get_highlighter(text, None),
                            raw_url=app.config['FILE_URL'].format(name=uf.name),
                        )))
                        objects.append(pb)

                uploaded_files.append((uf, pb))
            except FileTooLargeError as ex:
                num_bytes, = ex.args
                return jsonify({
                    'success': False,
                    'error': '{} ({}) exceeded the maximum file size limit of {}'.format(
                        f.filename,
                        human_size(num_bytes),
                        human_size(app.config['MAX_UPLOAD_SIZE']),
                    ),
                }), 413

        details_obj = ctx.enter_context(
            HtmlToStore.from_html(render_template(
                'details.html',
                uploads=uploaded_files,
            )),
        )
        objects.append(details_obj)

        upload_objects(objects)

    if 'json' in request.args:
        return jsonify({
            'success': True,
            'redirect': details_obj.url,
            'uploaded_files': {
                uf.human_name: {
                    'raw': uf.url,
                    'paste': pb.url if pb is not None else None,
                }
                for uf, pb in uploaded_files
            },
        })
    else:
        return redirect(details_obj.url)


@app.route('/paste', methods={'POST'})
def paste():
    """Paste and redirect."""
    text = request.form['text']

    with contextlib.ExitStack() as ctx:
        objects = []

        # Browsers always send \r\n for the pasted text, which leads to bad
        # newlines when curling the raw text (#28).
        try:
            uf = ctx.enter_context(UploadedFile.from_text(text.replace('\r\n', '\n')))
        except FileTooLargeError as ex:
            num_bytes, = ex.args
            return 'Exceeded the max upload size of {} (tried to paste {})'.format(
                human_size(app.config['MAX_UPLOAD_SIZE']),
                human_size(num_bytes),
            ), 413
        objects.append(uf)

        lang = request.form['language']
        if lang != 'rendered-markdown':
            paste_obj = ctx.enter_context(HtmlToStore.from_html(render_template(
                'paste.html',
                text=text,
                highlighter=get_highlighter(text, lang),
                raw_url=app.config['FILE_URL'].format(name=uf.name),
            )))
            objects.append(paste_obj)
        else:
            paste_obj = ctx.enter_context(HtmlToStore.from_html(render_template(
                'markdown.html',
                text=text,
                raw_url=app.config['FILE_URL'].format(name=uf.name),
            )))
            objects.append(paste_obj)

        upload_objects(objects)

    return redirect(paste_obj.url)
