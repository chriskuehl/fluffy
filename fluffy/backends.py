"""Backends handle storing of uploaded files. A backend should implement
__init__(self, options) where options will be the dict of options given in the
fluffy settings, and store(self, stored_file) which stores an uploaded file.

All other details are left up to your implementation.
"""
import os
import mimetypes

import boto3

from fluffy import app


class Backend:

    def __init__(self, options):
        self.options = options


class FileBackend(Backend):
    """Storage backend which stores files and info pages on the local disk."""

    def store(self, stored_file):
        path = self.options['file_path'].format(name=stored_file.name)
        info_path = self.options['info_path'].format(name=stored_file.name)

        try:
            # store the file itself
            print('Writing to {}...'.format(path))
            stored_file.file.save(path)

            # store the info page
            print('Writing info page to {}...'.format(info_path))
            with open(info_path, 'wb') as dest:
                dest.write(stored_file.info_html.encode('utf-8'))
        except IOError as e:
            internal = 'Received IOError: {}'.format(e)
            display = "Sorry, we weren't able to save your file."
            raise BackendException(internal, display)


class S3Backend(Backend):
    """Storage backend which uploads to S3 using boto3."""

    def store(self, stored_file):
        s3 = boto3.resource('s3')
        objects = [
            {
                'path': self.options['file_s3path'].format(name=stored_file.name),
                'body': stored_file.file,
                'bucket': self.options['file_bucket'],
            },
            {
                'path': self.options['info_s3path'].format(name=stored_file.name),
                'body': stored_file.info_html.encode('utf8'),
                'bucket': self.options['info_bucket'],
            },
        ]
        for obj in objects:
            mime, encoding = mimetypes.guess_type(obj['path'])
            if not mime:
                mime = 'applicaton/octet-stream'

            s3.Bucket(obj['bucket']).put_object(
                Key=obj['path'],
                Body=obj['body'],
                ContentType=mime,
            )


class BackendException(Exception):
    """Exception to be raised when a backend encounters an error trying to store
    a file. user_message will be displayed to the user, internal_message will
    be logged and not displayed.

    BackendException will display a "friendly" message to the user. All other
    uncaught exceptions will display a generic error.
    """

    def __init__(self, internal_message, display_message):
        self.display_message = display_message
        Exception.__init__(self, internal_message)

backends = {
    'file': FileBackend,
    's3': S3Backend,
}


def get_backend():
    """Returns a backend instance as configured in the settings."""
    conf = app.config['STORAGE_BACKEND']
    name, options = conf['name'], conf['options']

    return backends[name](options)
