"""File storage backends.

Backends are required to be able to store both HTML and objects. HTML should be
served as text/html, objects should be served as something safe.

Some backends can control the mimetype (S3), some can't (file). So be careful
what you do!
"""
import functools

import boto3

from fluffy import app


# TODO: FileBackend is broken
class FileBackend:
    """Storage backend which stores files and info pages on the local disk."""

    def store(self, stored_file):
        path = self.options['file_path'].format(name=stored_file.name)
        stored_file.file.save(path)


class S3Backend:
    """Storage backend which uploads to S3 using boto3."""

    def _store(self, obj):
        # S3 lets us specify mimetypes per file :D
        s3 = boto3.resource('s3')
        s3.Bucket(app.config['STORAGE_BACKEND']['bucket']).put_object(
            Key=app.config['STORAGE_BACKEND']['s3path'].format(name=obj.name),
            Body=obj.open_file,
            ContentType=obj.mimetype,
        )

    store_object = _store
    store_html = _store


@functools.lru_cache()
def get_backend():
    """Return current backend."""
    return {
        'file': FileBackend,
        's3': S3Backend,
    }[app.config['STORAGE_BACKEND']['name']]()
