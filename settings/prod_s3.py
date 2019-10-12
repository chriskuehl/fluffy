# fluffy-specific configuration options
# storage backend (how are the files stored after being uploaded?)
STORAGE_BACKEND = {
    "name": "s3",
    "bucket": "fluffy.cc",
    "s3path": "{name}",
    "asset_bucket": "fluffy.cc",
    "asset_s3path": "assets/{name}",
}

# branding to show in heading
BRANDING = "fluffy"

# URL patterns
HOME_URL = "https://fluffy.cc/"
FILE_URL = "https://i.fluffy.cc/{name}"
HTML_URL = "https://i.fluffy.cc/{name}"

STATIC_ASSETS_URL = "https://i.fluffy.cc/assets/{name}"

# abuse contact email address
ABUSE_CONTACT = "report@fluffy.cc"

# max upload size per request (in bytes)
MAX_UPLOAD_SIZE = 20 * 1048576

# max size Flask will accept; maybe a little larger?
MAX_CONTENT_LENGTH = MAX_UPLOAD_SIZE * 10

# file extensions to forbid for uploads
EXTENSION_BLACKLIST = frozenset(("exe",))
