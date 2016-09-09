# fluffy-specific configuration options
# storage backend (how are the files stored after being uploaded?)
STORAGE_BACKEND = {{
    'name': 'file',
    'object_path': '{object_path}',
    'html_path': '{html_path}',
}}

# branding to show in heading
BRANDING = 'fluffy'

# URL patterns
HOME_URL = '{home_url}'
FILE_URL = '{file_url}'
HTML_URL = '{html_url}'

STATIC_ASSETS_URL = '{static_assets_url}'

# abuse contact email address
ABUSE_CONTACT = 'abuse@example.com'

# max upload size per file (in bytes)
MAX_UPLOAD_SIZE = 10 * 1048576  # 10 MB

# max size Flask will accept; maybe a little larger?
MAX_CONTENT_LENGTH = MAX_UPLOAD_SIZE * 2
