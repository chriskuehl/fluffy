import io
import mimetypes
import os
import tempfile
from collections import namedtuple
from contextlib import contextmanager

from cached_property import cached_property

from fluffy.app import app
from fluffy.utils import content_is_binary
from fluffy.utils import gen_unique_id
from fluffy.utils import ONE_MB


MIME_WHITELIST = frozenset([
    'application/pdf',
    'application/x-ruby',
    'audio/',
    'image/',
    'text/plain',
    'text/x-python',
    'text/x-sh',
    'video/',
])


class ObjectToStore:

    @property
    def open_file(self):
        raise NotImplementedError()

    @property
    def mimetype(self):
        raise NotImplementedError()

    @property
    def name(self):
        raise NotImplementedError()


class UploadedFile(namedtuple('UploadedFile', (
        'human_name',
        'num_bytes',
        'open_file',
        'unique_id',
)), ObjectToStore):

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
                unique_id=gen_unique_id(),
            )

    @classmethod
    @contextmanager
    def from_text(cls, text):
        with io.BytesIO(text.encode('utf8')) as open_file:
            num_bytes = len(text)
            if num_bytes > app.config['MAX_UPLOAD_SIZE']:
                raise FileTooLargeError()

            yield cls(
                human_name='plaintext.txt',
                num_bytes=num_bytes,
                open_file=open_file,
                unique_id=gen_unique_id(),
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
    def probably_binary(self):
        p = content_is_binary(self.open_file.read(ONE_MB))
        self.open_file.seek(0)
        return p

    @cached_property
    def full_content(self):
        content = self.open_file.read()
        self.open_file.seek(0)
        return content

    @cached_property
    def mimetype(self):
        mime, _ = mimetypes.guess_type(self.name)
        if (
                mime and
                any(mime.startswith(check) for check in MIME_WHITELIST)
        ):
            return mime
        else:
            if self.probably_binary:
                return 'application/octet-stream'
            else:
                return 'text/plain'

    @cached_property
    def download_url(self):
        return app.config['FILE_URL'].format(name=self.name)


class HtmlToStore(namedtuple('HtmlToStore', (
    'name',
    'open_file',
)), ObjectToStore):

    @classmethod
    @contextmanager
    def from_html(cls, html):
        with io.BytesIO(html.encode('utf8')) as open_file:
            yield cls(
                name=gen_unique_id() + '.html',
                open_file=open_file,
            )

    @property
    def mimetype(self):
        return 'text/html'

    @cached_property
    def url(self):
        return app.config['HTML_URL'].format(name=self.name)


class FileTooLargeError(Exception):
    pass
