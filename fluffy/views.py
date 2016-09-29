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
    )


@app.route('/upload', methods={'POST'})
def upload():
    """Process an upload and return JSON status."""
    uploaded_files = []
    for f in request.files.getlist('file'):
        try:
            with UploadedFile.from_http_file(f) as uf:
                get_backend().store_object(uf)

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
                        with HtmlToStore.from_html(render_template(
                            'paste.html',
                            text=text,
                            highlighter=get_highlighter(text, None),
                            raw_url=app.config['FILE_URL'].format(name=uf.name),
                        )) as pb:
                            get_backend().store_html(pb)

            uploaded_files.append((uf, pb))
        except FileTooLargeError:
            return jsonify({
                'success': False,
                'error': '{} exceeded the maximum file size limit of {}'.format(
                    f.filename,
                    human_size(app.config['MAX_UPLOAD_SIZE']),
                ),
            })

    with HtmlToStore.from_html(render_template(
        'details.html',
        uploads=uploaded_files,
    )) as details_obj:
        get_backend().store_html(details_obj)

    url = app.config['HTML_URL'].format(name=details_obj.name)

    if 'json' in request.args:
        return jsonify({
            'success': True,
            'redirect': url,
        })
    else:
        return redirect(url)


@app.route('/paste', methods={'POST'})
def paste():
    """Paste and redirect."""
    text = request.form['text']

    # TODO: make this better
    assert 0 <= len(text) <= ONE_MB, len(text)

    with UploadedFile.from_text(text) as uf:
        get_backend().store_object(uf)

    lang = request.form['language']
    if lang != 'rendered-markdown':
        with HtmlToStore.from_html(render_template(
            'paste.html',
            text=text,
            highlighter=get_highlighter(text, lang),
            raw_url=app.config['FILE_URL'].format(name=uf.name),
        )) as paste_obj:
            get_backend().store_html(paste_obj)
    else:
        with HtmlToStore.from_html(render_template(
            'markdown.html',
            text=text,
            raw_url=app.config['FILE_URL'].format(name=uf.name),
        )) as paste_obj:
            get_backend().store_html(paste_obj)

    url = app.config['HTML_URL'].format(name=paste_obj.name)
    return redirect(url)
