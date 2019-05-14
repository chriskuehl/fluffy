import contextlib
import json
import time

from flask import jsonify
from flask import redirect
from flask import render_template
from flask import request

from fluffy import version
from fluffy.app import app
from fluffy.component.backends import get_backend
from fluffy.component.highlighting import get_highlighter
from fluffy.component.highlighting import UI_LANGUAGES_MAP
from fluffy.component.styles import STYLES_BY_CATEGORY
from fluffy.models import ExtensionForbiddenError
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


def upload_objects(objects, metadata_url=None):
    # TODO: make metadata_url mandatory (need to support it for uploads too)
    links = sorted(obj.url for obj in objects)
    for obj in objects:
        if isinstance(obj, HtmlToStore):
            get_backend().store_html(obj, links, metadata_url)
        else:
            get_backend().store_object(obj, links, metadata_url)


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
                            highlighter=get_highlighter(text, None, uf.human_name),
                            raw_url=app.config['FILE_URL'].format(name=uf.name),
                            styles=STYLES_BY_CATEGORY,
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
            except ExtensionForbiddenError as ex:
                extension, = ex.args
                return jsonify({
                    'success': False,
                    'error': 'Sorry, files with the extension ".{}" are not allowed.'.format(extension),
                }), 403

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
    # Browsers always send \r\n for the pasted text, which leads to bad
    # newlines when curling the raw text (#28).
    transformed_text = text.replace('\r\n', '\n')

    with contextlib.ExitStack() as ctx:
        objects = []

        # Raw text object
        try:
            uf = ctx.enter_context(UploadedFile.from_text(transformed_text))
        except FileTooLargeError as ex:
            num_bytes, = ex.args
            return 'Exceeded the max upload size of {} (tried to paste {})'.format(
                human_size(app.config['MAX_UPLOAD_SIZE']),
                human_size(num_bytes),
            ), 413
        objects.append(uf)

        # HTML view (Markdown or paste)
        lang = request.form['language']
        if lang != 'rendered-markdown':
            highlighter = get_highlighter(text, lang, None)
            lang_title = highlighter.name
            paste_obj = ctx.enter_context(HtmlToStore.from_html(render_template(
                'paste.html',
                text=text,
                highlighter=highlighter,
                raw_url=app.config['FILE_URL'].format(name=uf.name),
                styles=STYLES_BY_CATEGORY,
            )))
            objects.append(paste_obj)
        else:
            lang_title = 'Rendered Markdown'
            paste_obj = ctx.enter_context(HtmlToStore.from_html(render_template(
                'markdown.html',
                text=text,
                raw_url=app.config['FILE_URL'].format(name=uf.name),
            )))
            objects.append(paste_obj)

        # Metadata JSON object
        metadata = {
            'server_version': version,
            'uploaded_files': {
                'html': paste_obj.url,
                'raw': uf.url,
            },
            'timestamp': time.time(),
            'upload_type': 'paste',
            'paste_details': {
                'language': {
                    'title': lang_title,
                },
                'num_lines': len(transformed_text.splitlines()),
                'raw_text': transformed_text,
            },
        }
        metadata_obj = ctx.enter_context(UploadedFile.from_text(
            json.dumps(metadata, indent=4, sort_keys=True),
            human_name='metadata.json',
        ))
        objects.append(metadata_obj)

        upload_objects(objects, metadata_url=metadata_obj.url)

    if 'json' in request.args:
        return jsonify({
            'success': True,
            'redirect': paste_obj.url,
            'uploaded_files': {
                'paste': {
                    'raw': uf.url,
                    'paste': paste_obj.url,
                    'metadata': metadata_obj.url,
                },
            },
        })
    else:
        return redirect(paste_obj.url)
