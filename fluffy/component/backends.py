"""File storage backends.

Backends are required to be able to store both HTML and objects. HTML should be
served as text/html, objects should be served as something safe.

Some backends can control the mimetype (S3), some can't (file). So be careful
what you do!
"""
import functools
import shutil
import typing

import boto3

from fluffy.app import app
from fluffy.models import HtmlToStore
from fluffy.models import UploadedFile


class FileBackend:
    """Storage backend which stores files and info pages on the local disk."""

    def _store(
        self,
        path_key: str,
        obj: typing.Union[HtmlToStore, UploadedFile],
    ):
        path = app.config['STORAGE_BACKEND'][path_key].format(name=obj.name)
        with open(path, 'wb') as f:
            shutil.copyfileobj(obj.open_file, f)
            obj.open_file.seek(0)

    # TODO: support links for file backend somehow?
    def store_object(
        self,
        obj: UploadedFile,
        links: typing.Sequence[str],
        metadata_url: str,
    ) -> None:
        self._store('object_path', obj)

    def store_html(
        self,
        obj: HtmlToStore,
        links: typing.Sequence[str],
        metadata_url: str,
    ) -> None:
        self._store('html_path', obj)


class S3Backend:
    """Storage backend which uploads to S3 using boto3."""

    def _store(
        self,
        obj: typing.Union[HtmlToStore, UploadedFile],
        links: typing.Sequence[str],
        metadata_url: str,
    ) -> None:
        # We always use a new session in case the keys have been rotated on disk.
        session = boto3.session.Session()
        s3 = session.resource('s3')
        s3.Bucket(app.config['STORAGE_BACKEND']['bucket']).put_object(
            Key=app.config['STORAGE_BACKEND']['s3path'].format(name=obj.name),
            Body=obj.open_file,
            ContentType=obj.mimetype,
            Metadata={
                'fluffy-links': '; '.join(links),
                'fluffy-metadata': metadata_url,
            },
            ContentDisposition=obj.content_disposition_header,
            # Allow the bucket owner to control the object, for cases where the
            # bucket is owned by a different account.
            ACL='bucket-owner-full-control',
        )
        obj.open_file.seek(0)

    # S3 lets us specify mimetypes per file :D
    store_object = _store
    store_html = _store


@functools.lru_cache()
def get_backend():
    """Return current backend."""
    return {
        'file': FileBackend,
        's3': S3Backend,
    }[app.config['STORAGE_BACKEND']['name']]()
