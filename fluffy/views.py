import json

from flask import jsonify
from flask import redirect
from flask import render_template
from flask import request
from flask import url_for

from fluffy import app
from fluffy.backends import get_backend
from fluffy.models import FileTooLargeError
from fluffy.models import UploadedFile
from fluffy.utils import decode
from fluffy.utils import encode
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
                get_backend().store(uf)
            uploaded_files.append(uf)
        except FileTooLargeError:
            return jsonify({
                'success': False,
                'error': '{} exceeded the maximum file size limit of {}'.format(
                    f.filename,
                    human_size(app.config['MAX_UPLOAD_SIZE']),
                ),
            })

    url = url_for(
        'details',
        enc=encode(json.dumps([uf.serialized for uf in uploaded_files])),
    )
    if 'json' in request.args:
        return jsonify({
            'success': True,
            'redirect': url,
        })
    else:
        return redirect(url)


@app.route('/details/<enc>')
def details(enc):
    """Displays details about an upload (or any set of files, really).

    enc is the encoded list of detail tuples, as returned by get_details.
    """
    return render_template(
        'details.html',
        uploads=[UploadedFile.deserialized(uf) for uf in json.loads(decode(enc))],
    )


# TODO: remove this
@app.route('/test/paste')
def paste():
    if not app.debug:
        return
    import requests
    text = requests.get('https://raw.githubusercontent.com/ocf/ocfweb/master/ocfweb/account/register.py').text
    return render_template(
        'paste.html',
        num_lines=len(text.splitlines()),
        text=text,
    )


def get_client_ip():
    # TODO: improve this to better handle proxies
    return request.remote_addr
