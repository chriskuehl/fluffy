from flask import jsonify
from flask import redirect
from flask import render_template
from flask import request

import fluffy.highlighting  # noqa
from fluffy import app
from fluffy.backends import get_backend
from fluffy.models import FileTooLargeError
from fluffy.models import HtmlToStore
from fluffy.models import UploadedFile
from fluffy.utils import human_size


@app.route('/')
def index():
    return render_template('index.html')


@app.route('/upload', methods={'POST'})
def upload():
    """Process an upload and return JSON status."""
    uploaded_files = []
    for f in request.files.getlist('file'):
        try:
            with UploadedFile.from_http_file(f) as uf:
                get_backend().store_object(uf)
            uploaded_files.append(uf)
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


# TODO: remove this
@app.route('/test/paste')
def paste():
    import requests
    text = requests.get('https://raw.githubusercontent.com/ocf/ocfweb/master/ocfweb/account/register.py').text
    return render_template(
        'paste.html',
        num_lines=len(text.splitlines()),
        text=text,
    )
