import functools
import os
import re
import subprocess
from pathlib import Path

from flask import url_for

from fluffy.app import app


PROJECT_ROOT = Path(__file__).parent.parent
STATIC_ROOT = PROJECT_ROOT / 'static'


@functools.lru_cache(10000)
def hash_for_asset(path):
    asset_hash_path = STATIC_ROOT / (path + '.hash')
    with asset_hash_path.open('r') as f:
        return f.read().strip()


def name_for_asset(path):
    fname = re.sub('[^a-zA-Z0-9]', '-', path)
    return '{}-{}.{}'.format(
        fname,
        hash_for_asset(path),
        os.path.splitext(path)[1].lstrip('.'),
    )


def asset_url(path):
    """Return a URL for the asset, suitable for dev or prod.

    In dev, we use Flask to serve it.
    In prod, we calculate the SHA and try to serve it from the specified CDN.

    The purpose of calculating hashes and never removing old ones is that the
    stored files (info pages) want to reference a specific version of the
    assets. We don't want old info pages to break if we update with new
    incompatible styles.
    """
    if app.debug:
        return url_for('static', filename=path, _external=True)
    else:
        return app.config['STATIC_ASSETS_URL'].format(
            name=name_for_asset(path),
        )


def upload_assets():
    """Upload assets. Currently supports only S3."""
    commands = []
    for root, dirs, files in os.walk(str(STATIC_ROOT)):
        for fname in files:
            if not fname.endswith('.hash'):
                continue
            if '.debug.' in fname:
                continue

            asset_hash_path = os.path.join(root, fname)
            asset_path = asset_hash_path[:-5]

            if os.path.isfile(asset_path):
                commands.append('aws s3 cp {} s3://{}/{}'.format(
                    asset_path,
                    app.config['STORAGE_BACKEND']['asset_bucket'],
                    app.config['STORAGE_BACKEND']['asset_s3path'].format(
                        name=name_for_asset(
                            os.path.relpath(asset_path, str(STATIC_ROOT)),
                        ),
                    ),
                ))

    print('=' * 50)
    print('\n'.join(commands))
    print('=' * 50)
    if input('Want me to run these for you? [yN] ').lower() in ('y', 'yes'):
        for command in commands:
            subprocess.check_call(command, shell=True)
