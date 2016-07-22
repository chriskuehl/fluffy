import logging
import sys

from flask import Flask


app = Flask(__name__)
app.config.from_envvar('FLUFFY_SETTINGS')
app.logger.addHandler(logging.StreamHandler(sys.stderr))
app.logger.setLevel(logging.DEBUG)


@app.context_processor
def defaults():
    from fluffy.assets import asset_url as real_asset_url
    return {
        'abuse_contact': app.config['ABUSE_CONTACT'],
        'allow_debug': True,
        'asset_url': real_asset_url,
        'branding': app.config['BRANDING'],
        'home_url': app.config['HOME_URL'],
        'num_lines': lambda text: len(text.splitlines()),
    }
