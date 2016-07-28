import os
import shutil
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path

import ephemeral_port_reserve
import pytest
import requests


PROJECT_ROOT = Path(__file__).parent.parent
TESTING_DIR = PROJECT_ROOT / 'testing'


def _templated_config(tempdir, app_port, static_port):
    with (PROJECT_ROOT / 'settings' / 'test_files.py').open('r') as f:
        return f.read().format(
            server_name='localhost:{}'.format(app_port),
            object_path=os.path.join(tempdir, 'object', '{name}'),
            html_path=os.path.join(tempdir, 'html', '{name}'),
            home_url='http://localhost:{}/'.format(app_port),
            file_url='http://localhost:{}/object/{{name}}'.format(static_port),
            html_url='http://localhost:{}/html/{{name}}'.format(static_port),
            static_assets_url='http://localhost:{}/html/{{name}}'.format(static_port),
        )


def _wait_for_http(url):
    for _ in range(500):
        try:
            req = requests.get(url)
        except requests.exceptions.ConnectionError:  # pragma: no cover
            pass
        else:
            if req.status_code == 200:
                break
        time.sleep(0.01)  # pragma: no cover
    else:  # pragma: no cover
        raise RuntimeError('Timed out trying to access: {}'.format(url))


@pytest.yield_fixture(scope='session')
def running_server():
    """A running fluffy server.

    Starts an app server on one port, and an http.server on another port to
    serve the static files (much like pgctl does in dev).
    """
    tempdir = tempfile.mkdtemp()

    os.mkdir(os.path.join(tempdir, 'object'))
    os.mkdir(os.path.join(tempdir, 'html'))

    app_port = ephemeral_port_reserve.reserve()
    static_port = ephemeral_port_reserve.reserve()

    settings_path = os.path.join(tempdir, 'settings.py')
    with open(settings_path, 'w') as f:
        f.write(_templated_config(tempdir, app_port, static_port))

    os.environ['FLUFFY_SETTINGS'] = settings_path
    app_server = subprocess.Popen(
        (
            sys.executable,
            '-m', 'gunicorn.app.wsgiapp',
            '-b', '127.0.0.1:{}'.format(app_port),
            'fluffy.run:app',
        ),
        env={
            'COVERAGE_PROCESS_START': os.environ.get('COVERAGE_PROCESS_START', ''),
            'FLUFFY_SETTINGS': settings_path,
        },
    )
    static_server = subprocess.Popen(
        (
            sys.executable,
            '-m', 'http.server',
            '--bind', '127.0.0.1',
            str(static_port),
        ),
        cwd=tempdir,
    )

    _wait_for_http('http://localhost:{}'.format(app_port))
    _wait_for_http('http://localhost:{}'.format(static_port))

    yield {
        'home': 'http://localhost:{}'.format(app_port),
    }

    time.sleep(1)

    static_server.send_signal(signal.SIGTERM)
    assert static_server.wait() == -signal.SIGTERM, static_server.returncode

    app_server.send_signal(signal.SIGTERM)
    assert app_server.wait() == 0, app_server.returncode

    shutil.rmtree(tempdir)
