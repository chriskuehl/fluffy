import io
import mimetypes
import os
import tempfile
import urllib.parse
from collections import namedtuple
from contextlib import contextmanager

from cached_property import cached_property
from identify import identify

from fluffy.app import app
from fluffy.utils import gen_unique_id


# Mime types which are allowed to be presented as detected.
# TODO: I think we actually only need to prevent text/html (and any HTML
# variants like XHTML)?
MIME_WHITELIST = (
    'application/javascript',
    'application/json',
    'application/pdf',
    'application/x-ruby',
    'audio/',
    'image/',
    'text/css',
    'text/plain',
    'text/x-python',
    'text/x-sh',
    'video/',
)

# Mime types which should be displayed inline in the browser, as opposed to
# being downloaded. This is used to populate the Content-Disposition header.
# Only binary MIMEs need to be whitelisted here, since detected non-binary
# files are always inline.
INLINE_DISPLAY_MIME_WHITELIST = (
    'application/pdf',
    'audio/',
    'image/',
    'video/',
)


class ObjectToStore:

    @property
    def open_file(self):
        raise NotImplementedError()

    @property
    def mimetype(self):
        raise NotImplementedError()

    @cached_property
    def content_disposition_header(self):
        raise NotImplementedError()

    @property
    def name(self):
        raise NotImplementedError()


class UploadedFile(
    namedtuple(
        'UploadedFile',
        (
            'human_name',
            'num_bytes',
            'open_file',
            'unique_id',
        ),
    ),
    ObjectToStore,
):

    def __new__(cls, *args, **kwargs):
        uf = super().__new__(cls, *args, **kwargs)
        if uf.extension in app.config.get('EXTENSION_BLACKLIST', ()):
            raise ExtensionForbiddenError(uf.extension)
        else:
            return uf

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
                raise FileTooLargeError(num_bytes)
            tf.seek(0)

            yield cls(
                human_name=f.filename,
                num_bytes=num_bytes,
                open_file=tf,
                unique_id=gen_unique_id(),
            )

    @classmethod
    @contextmanager
    def from_text(cls, text, human_name='plaintext.txt'):
        with io.BytesIO(text.encode('utf8')) as open_file:
            num_bytes = len(text)
            if num_bytes > app.config['MAX_UPLOAD_SIZE']:
                raise FileTooLargeError(num_bytes)

            yield cls(
                human_name=human_name,
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
        p = not identify.is_text(self.open_file)
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
        if mime and mime.startswith(MIME_WHITELIST):
            return mime
        else:
            if self.probably_binary:
                return 'application/octet-stream'
            else:
                return 'text/plain'

    @cached_property
    def content_disposition_header(self):
        if self.mimetype.startswith(INLINE_DISPLAY_MIME_WHITELIST) or not self.probably_binary:
            render_type = 'inline'
        else:
            render_type = 'attachment'
        return '{}; filename="{}"; filename*=utf-8\'\'{}'.format(
            render_type,
            self.human_name.replace('"', ''),
            urllib.parse.quote(self.human_name, encoding='utf-8'),
        )

    @cached_property
    def url(self):
        return app.config['FILE_URL'].format(name=self.name)


class HtmlToStore(
    namedtuple(
        'HtmlToStore',
        (
            'name',
            'open_file',
        ),
    ),
    ObjectToStore,
):

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

    @property
    def content_disposition_header(self):
        # inline => render as HTML as opposed to downloading the HTML
        return 'inline'

    @cached_property
    def url(self):
        return app.config['HTML_URL'].format(name=self.name)


class FileTooLargeError(Exception):
    pass


class ExtensionForbiddenError(Exception):
    pass
