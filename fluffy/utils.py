import base64
import os
import zlib

ONE_GB = 1073741824
ONE_MB = 1048576
ONE_KB = 1024


def trusted_network(ip):
    """Returns whether a given address is a member of a trusted network."""
    # TODO: this
    return False
    # addr = IPAddress(ip)
    # return any((addr in IPNetwork(net) for net in app.config['TRUSTED_NETWORKS']))


def human_size(size):
    if size / ONE_GB >= 1:
        return '{:.1f} GB'.format(size / ONE_GB)
    elif size / ONE_MB >= 1:
        return '{:.1f} MB'.format(size / ONE_MB)
    elif size / ONE_KB >= 1:
        return '{:.1f} KB'.format(size / ONE_KB)
    else:
        return '{} bytes'.format(size)

# constants to represent zipped or unzipped (at start of encoded string)
NOT_ZIPPED = 'u'
ZIPPED = 'z'


def encode(plaintext):
    """Encodes a string for inclusion in the query string.

    Works internally by serializing the object into JSON, gzipping the JSON,
    and then base64 encoding it.

    The serialized data will only be zipped if doing so would result in a
    shorter string. The returned string will start with ZIPPED or NOT_ZIPPED
    (a single character constant) to indicate zipped or unzipped.
    """
    def encode(s):
        return base64.urlsafe_b64encode(s).decode('utf8')

    zipped = zlib.compress(plaintext.encode('utf8'))
    choices = (
        NOT_ZIPPED + encode(plaintext.encode('utf-8')),
        ZIPPED + encode(zipped),
    )
    return min(choices, key=len)


def decode(encoded):
    """Decodes a string encoded by encode_obj into an object."""
    zipped = encoded.startswith(ZIPPED)
    the_bytes = base64.urlsafe_b64decode(encoded[1:].encode('utf-8'))
    if zipped:
        return zlib.decompress(the_bytes).decode('utf-8')
    else:
        return the_bytes.decode('utf-8')


ICON_EXTENSIONS = [
    '7z', 'ai', 'bmp', 'doc', 'docx', 'gif', 'gz', 'html',
    'jpeg', 'jpg', 'midi', 'mp3', 'odf', 'odt', 'pdf', 'png', 'psd', 'rar',
    'rtf', 'svg', 'tar', 'txt', 'wav', 'xls', 'zip'
]


def get_extension_icon(extension):
    """Returns the filename of the icon to use for an extension, excluding
    the .png at the end of the icon name.
    """
    extension = extension.lower()
    return extension if extension in ICON_EXTENSIONS else 'unknown'


def trim_filename(name, length):
    """Trim a filename down to a desired maximum length, making attempts to
    preserve the important parts of the name.

    We prefer to preserve, in order:
    1. the extension in full
    2. the first three and last two characters of the name
    3. everything else, prioritizing the first characters
    """
    if len(name) <= length:
        return name

    name, ext = os.path.splitext(name)
    length = max(length, len(ext) + 5)

    if len(name) + len(ext) <= length + 2:
        # we can't really do better
        return name + ext

    length += 1  # don't count the dot in the extension
    length += 1  # count elipses as two characters

    prefix = ''
    suffix = name[-2:]

    def get_result():
        return prefix.strip() + '...' + suffix.strip() + ext

    for i in range(3, len(name) - 2):
        prefix = name[:i]
        result = get_result()

        if len(result) >= length:
            return result

    return get_result()
