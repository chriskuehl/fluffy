import os
import time

from flask import jsonify
from flask import redirect
from flask import render_template
from flask import request
from flask import url_for

from fluffy import app
from fluffy.backends import get_backend
from fluffy.models import StoredFile
from fluffy.utils import decode_obj
from fluffy.utils import encode_obj
from fluffy.utils import get_extension_icon
from fluffy.utils import trim_filename
from fluffy.utils import trusted_network
from fluffy.utils import validate_files
from fluffy.utils import ValidationException


@app.route('/')
def index():
    trusted = trusted_network(get_client_ip())
    return render_template('index.html', trusted=trusted)


@app.route('/upload', methods={'POST'})
def upload():
    """Process an upload and return JSON status."""
    try:
        backend = get_backend()
        trusted_user = trusted_network(get_client_ip())
        file_list = request.files.getlist('file')

        validate_files(file_list, trusted_user)

        stored_files = [StoredFile(file) for file in file_list]

        start = time.time()
        print('Storing {} files...'.format(len(stored_files)))

        for stored_file in stored_files:
            print('Storing {}...'.format(stored_file.name))
            backend.store(stored_file)

        elapsed = time.time() - start
        print('Stored {} files in {:.1f} seconds.'.format(len(stored_files), elapsed))

        # redirect to the details page if multiple files were uploaded
        # otherwise, redirect to the info page of the single file
        if len(stored_files) > 1:
            details = [get_details(f) for f in stored_files]
            url = url_for('details', enc=encode_obj(details))
        else:
            url = app.config['INFO_URL'].format(name=stored_files[0].name)

        if 'json' in request.args:
            return jsonify({
                'success': True,
                'redirect': url,
            })
        else:
            return redirect(url)
    except ValidationException as ex:
        print('Refusing to accept file (failed validation):')
        print('\t{}'.format(ex))

        return jsonify({
            'success': False,
            'error': str(ex),
        })


def get_details(stored_file):
    """Returns a tuple of details of a single stored file to be included in the
    parameters of the info page.

    Details in the tuple:
      - stored name
      - human name without extension (to save space)
    """
    human_name = os.path.splitext(stored_file.file.name)[0]

    return (stored_file.name, human_name)


@app.route('/details/<enc>')
def details(enc):
    """Displays details about an upload (or any set of files, really).

    enc is the encoded list of detail tuples, as returned by get_details.
    """
    req_details = decode_obj(enc)
    details = [get_full_details(file) for file in req_details]

    return render_template('details.html', details=details)


def get_full_details(file):
    """Returns a dictionary of details for a file given a detail tuple."""
    stored_name = file[0]
    name = file[1]  # original file name
    ext = os.path.splitext(stored_name)[1]

    return {
        'download_url': app.config['FILE_URL'].format(name=stored_name),
        'info_url': app.config['INFO_URL'].format(name=stored_name),
        'name': trim_filename(name + ext, 17),  # original name is stored w/o extension
        'extension': get_extension_icon(ext[1:] if ext else '')
    }


def get_client_ip():
    # TODO: improve this to better handle proxies
    return request.remote_addr
