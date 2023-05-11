import functools
import logging
import os
import sys

import markupsafe
from flask import Flask

from fluffy import version


app = Flask(__name__)
app.config.from_envvar('FLUFFY_SETTINGS')
app.logger.addHandler(logging.StreamHandler(sys.stderr))
app.logger.setLevel(logging.DEBUG)


def _cached_if_prod(fn):
    cached_fn = functools.lru_cache(fn)

    @functools.wraps(fn)
    def wrapped(*args, **kwargs):
        # Note that `app.debug` isn't correct at import time so this needs to
        # be done at the actual call time.
        if app.debug:
            return fn(*args, **kwargs)
        else:
            return cached_fn(*args, **kwargs)

    return wrapped


@_cached_if_prod
def _inline_js(path):
    assert '..' not in path, path
    with open(os.path.join(os.path.dirname(__file__), os.path.join('static/js', path))) as f:
        return markupsafe.Markup(f'<script>\n{f.read()}\n</script>')


@app.context_processor
def defaults():
    from fluffy.component.assets import asset_url as real_asset_url
    return {
        'abuse_contact': app.config['ABUSE_CONTACT'],
        'app': app,
        'asset_url': real_asset_url,
        'branding': app.config['BRANDING'],
        'fluffy_version': version,
        'home_url': app.config['HOME_URL'],
        'custom_footer_html': app.config.get('CUSTOM_FOOTER_HTML'),
        'num_lines': lambda text: len(text.splitlines()),
        'inline_js': _inline_js,
    }
