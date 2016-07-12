"""File storage backends.

Backends know how to store an uploaded file, and not much else.

A backend is required to implement the interface:

    __init__(self, options)

        ...where options will be the dict of options given in the fluffy
        settings

    store(self, stored_file)

        ...which stores an uploaded file.

All other details are left up to your implementation.
"""
import boto3

from fluffy import app


class Backend:

    def __init__(self, options):
        self.options = options


class FileBackend(Backend):
    """Storage backend which stores files and info pages on the local disk."""

    def store(self, stored_file):
        path = self.options['file_path'].format(name=stored_file.name)
        print('Writing to {}...'.format(path))
        stored_file.file.save(path)


class S3Backend(Backend):
    """Storage backend which uploads to S3 using boto3."""

    def store(self, upload):
        s3 = boto3.resource('s3')
        s3.Bucket(self.options['file_bucket']).put_object(
            Key=self.options['file_s3path'].format(name=upload.name),
            Body=upload.open_file,
            ContentType=upload.mimetype,
        )


backends = {
    'file': FileBackend,
    's3': S3Backend,
}


def get_backend():
    """Returns a backend instance as configured in the settings."""
    conf = app.config['STORAGE_BACKEND']
    name, options = conf['name'], conf['options']

    return backends[name](options)
