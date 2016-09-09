# fluffy-specific configuration options
# storage backend (how are the files stored after being uploaded?)
STORAGE_BACKEND = {
    'name': 's3',

    'bucket': 'fluffy.cc',
    's3path': '{name}',

    'asset_bucket': 'fluffy.cc',
    'asset_s3path': 'assets/{name}',
}

# branding to show in heading
BRANDING = 'fluffy'

# URL patterns
HOME_URL = 'http://localhost:5000/'
FILE_URL = 'https://i.fluffy.cc/{name}'
HTML_URL = 'https://i.fluffy.cc/{name}'

STATIC_ASSETS_URL = 'http://localhost:5000/{name}'

# abuse contact email address
ABUSE_CONTACT = 'abuse@example.com'

# max upload size per file (in bytes)
MAX_UPLOAD_SIZE = 10 * 1048576  # 10 MB

# max size Flask will accept; maybe a little larger?
MAX_CONTENT_LENGTH = MAX_UPLOAD_SIZE * 2
