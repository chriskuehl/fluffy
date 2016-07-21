import json
import mimetypes
import os
import random
import tempfile
from collections import namedtuple
from contextlib import contextmanager

from cached_property import cached_property

from fluffy import app


STORED_FILE_NAME_LENGTH = 32
STORED_FILE_NAME_CHARS = 'bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789'


class UploadedFile(namedtuple('UploadedFile', (
        'human_name',
        'num_bytes',
        'open_file',
        'unique_id',
))):

    @classmethod
    @contextmanager
    def from_http_file(cls, f):
        with tempfile.NamedTemporaryFile() as tf:
            # We don't know the file size until we start to save the file (the
            # client can lie about the uploaded size, and some browsers don't
            # even send it).
            f.save(tf)
            num_bytes = f.tell()
            if num_bytes > app.config['MAX_UPLOAD_SIZE']:
                raise FileTooLargeError()
            tf.seek(0)

            yield cls(
                human_name=f.filename,
                num_bytes=num_bytes,
                open_file=tf,
                unique_id=''.join(
                    random.choice(STORED_FILE_NAME_CHARS)
                    for _ in range(STORED_FILE_NAME_LENGTH)
                ),
            )

    @cached_property
    def name(self):
        """File name that will be stored."""
        if self.extension:
            return '{self.unique_id}.{self.extension}'.format(self=self)
        else:
            return self.unique_id

    @cached_property
    def extension(self):
        """Return file extension, or empty string."""
        _, ext = os.path.splitext(self.human_name)
        if ext.startswith('.'):
            ext = ext[1:]
        return ext

    @cached_property
    def mimetype(self):
        mime, _ = mimetypes.guess_type(self.name)
        if mime and mime.startswith('image/'):
            return mime
        else:
            return 'applicaton/octet-stream'

    @cached_property
    def download_url(self):
        return app.config['FILE_URL'].format(name=self.name)

    @classmethod
    def deserialized(cls, s):
        obj = json.loads(s)
        return cls(
            human_name=obj['human_name'],
            num_bytes=obj['num_bytes'],
            unique_id=obj['unique_id'],
            open_file=None,
        )

    @cached_property
    def serialized(self):
        return json.dumps({
            'human_name': self.human_name,
            'num_bytes': self.num_bytes,
            'unique_id': self.unique_id,
        })


class FileTooLargeError(Exception):
    pass
