# fluffy-specific configuration options
# storage backend (how are the files stored after being uploaded?)
# File backend
STORAGE_BACKEND = {
    'name': 'file',
    'options': {
        # use {name} as a placeholder for the file name
        'file_path': '/tmp/{name}',
        'info_path': '/tmp/{name}.html',
    }
}

# S3 CLI backend
STORAGE_BACKEND = {
    'name': 's3',
    'options': {
        # use {name} as a placeholder for the file name
        'file_name': '{name}',
        'info_name': '{name}.html',

        'file_bucket': 'fluffy.cc',
        'info_bucket': 'fluffy.cc',
        'file_s3path': '{name}',
        'info_s3path': 'info/{name}.html',

        'tmp_path': '/tmp/{name}',
    }
}

# URL patterns
HOME_URL = 'http://fluffy.cc/'

# use {name} as a placeholder for the file name
FILE_URL = 'http://i.fluffy.cc/{name}'
INFO_URL = 'http://i.fluffy.cc/info/{name}.html'

# abuse contact email address
ABUSE_CONTACT = 'abuse@example.com'

# max upload size per file (in bytes)
MAX_UPLOAD_SIZE = 10 * 1048576  # 10 MB

# max size Flask will accept; maybe a little larger?
# (could be much larger, if you let trusted users have no limit uploads)
MAX_CONTENT_LENGTH = MAX_UPLOAD_SIZE * 2

# trusted networks (requires netaddr package)
# prefix must be specified, for a single IPv4 use /32
TRUSTED_NETWORKS = (
    '127.0.0.0/8',  # loopback
    '172.16.0.0/16',
)

# do trusted networks get to violate max file size?
TRUSTED_NOMAXSIZE = True
