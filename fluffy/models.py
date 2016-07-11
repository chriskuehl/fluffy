import os
import random
from datetime import date

from flask import render_template

from fluffy import app
from fluffy.utils import get_extension_icon
from fluffy.utils import get_human_size
from fluffy.utils import trim_filename


class StoredFile:
    """A File object wraps an actual file and has a unique ID."""

    NAME_LENGTH = 32
    NAME_CHARS = 'bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789'

    def __init__(self, file):
        self.file = file
        self._generate_name()

    def _generate_name(self):
        """Generates a unique name for the file. We don't actually verify that
        the name is unique, but chances are very slim that it won't be."""

        name = ''.join(random.choice(StoredFile.NAME_CHARS)
                       for _ in range(StoredFile.NAME_LENGTH))

        extension = self.extension

        if extension:
            name += '.' + extension

        self.name = name

    @property
    def info_html(self):
        """Returns the HTML of the info page."""
        extension = self.extension

        params = {
            'name': trim_filename(self.file.filename, 17),
            # TODO: fix size
            'size': get_human_size(0),  # self.file.size),
            'date': date.today().strftime('%B %e, %Y'),
            'extension': get_extension_icon(extension),
            'download_url': app.config['FILE_URL'].format(name=self.name),
            'home_url': app.config['HOME_URL'],
        }

        return render_template('info.html', **params)

    @property
    def extension(self):
        """Returns extension without leading period, or empty string if no
        extension."""
        ext = os.path.splitext(self.file.filename)[1]
        return ext[1:] if ext else ''
