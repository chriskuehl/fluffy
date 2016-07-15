import base64
import os
import zlib

from fluffy import app

ONE_KB = 2**10
ONE_MB = 2**20
ONE_GB = 2**30


def trusted_network(ip):
    """Returns whether a given address is a member of a trusted network."""
    # TODO: this
    return False
    # addr = IPAddress(ip)
    # return any((addr in IPNetwork(net) for net in app.config['TRUSTED_NETWORKS']))


@app.template_filter()
def human_size(size):
    if size >= ONE_GB:
        return '{:.1f} GiB'.format(size / ONE_GB)
    elif size >= ONE_MB:
        return '{:.1f} MiB'.format(size / ONE_MB)
    elif size >= ONE_KB:
        return '{:.1f} KiB'.format(size / ONE_KB)
    else:
        return '{} bytes'.format(size)


def encode(plaintext):
    """Encodes a string for inclusion in the query string.

    Works by gzipping the string, and then urlsafe-base64 encoding it.
    """
    return base64.urlsafe_b64encode(
        zlib.compress(plaintext.encode('utf8')),
    ).decode('utf8')


def decode(encoded):
    """Decodes a string encoded by encode."""
    return zlib.decompress(
        base64.urlsafe_b64decode(encoded.encode('utf8')),
    ).decode('utf8')


# TODO: read these out of the package
ICON_EXTENSIONS = frozenset([
    '7z', 'ai', 'bmp', 'doc', 'docx', 'gif', 'gz', 'html',
    'jpeg', 'jpg', 'midi', 'mp3', 'odf', 'odt', 'pdf', 'png', 'psd', 'rar',
    'rtf', 'svg', 'tar', 'txt', 'wav', 'xls', 'zip'
])


@app.template_filter()
def icon_for_extension(extension):
    """Returns the filename of the icon to use for an extension, excluding
    the .png at the end of the icon name.
    """
    extension = extension.lower()
    if extension in ICON_EXTENSIONS:
        return extension
    else:
        return 'unknown'


@app.template_filter()
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
