import logging
import sys

from flask import Flask


app = Flask(__name__)
app.config.from_envvar('FLUFFY_SETTINGS')
app.logger.addHandler(logging.StreamHandler(sys.stderr))
app.logger.setLevel(logging.DEBUG)


@app.context_processor
def home_url():
    return {'home_url': app.config['HOME_URL']}


@app.context_processor
def asset_url():
    from fluffy.assets import asset_url as real_asset_url
    return {'asset_url': real_asset_url}
