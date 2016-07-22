# fluffy-specific configuration options
# storage backend (how are the files stored after being uploaded?)
# File backend
STORAGE_BACKEND = {
    'name': 'file',

    # use {name} as a placeholder for the file name
    'file_path': '/tmp/{name}',
    'info_path': '/tmp/{name}.html',
}

# S3 CLI backend
STORAGE_BACKEND = {
    'name': 's3',

    'bucket': 'fluffy.cc',
    's3path': '{name}',

    'asset_bucket': 'fluffy.cc',
    'asset_s3path': 'assets/{name}',
}

# URL patterns
HOME_URL = 'http://localhost:5000/'
FILE_URL = 'https://i.fluffy.cc/{name}'
HTML_URL = 'https://i.fluffy.cc/{name}'

STATIC_ASSETS_URL = 'https://i.fluffy.cc/assets/{name}'

# abuse contact email address
ABUSE_CONTACT = 'abuse@example.com'

# max upload size per file (in bytes)
MAX_UPLOAD_SIZE = 10 * 1048576  # 10 MB

# max size Flask will accept; maybe a little larger?
MAX_CONTENT_LENGTH = MAX_UPLOAD_SIZE * 2
