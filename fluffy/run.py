import os.path

from flask import render_template

from fluffy import app
from fluffy.highlighting import guess_lexer

# import views so the decorators run
import fluffy.views  # noreorder # noqa


def debug():

    @app.route('/test/paste')
    def view_paste():
        return render_template(
            'paste.html',
            text=open(os.path.join(os.path.dirname(__file__), 'models.py')).read(),
            lexer=guess_lexer('', 'python'),
            edit_url='#edit',
            raw_url='#raw',
        )

    app.run(debug=True)


if __name__ == '__main__':
    exit(debug())
