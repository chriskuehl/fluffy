import functools
import logging
import os
import sys

from flask import Flask

from fluffy import version


app = Flask(__name__)
app.config.from_envvar('FLUFFY_SETTINGS')
app.logger.addHandler(logging.StreamHandler(sys.stderr))
app.logger.setLevel(logging.DEBUG)


@functools.lru_cache
def _base_javascript():
    with open(os.path.join(os.path.dirname(__file__), 'static/js/base.js')) as f:
        return f.read()


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
        'base_javascript': _base_javascript(),
    }
