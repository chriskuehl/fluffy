import os
import random

from fluffy.app import app

ONE_KB = 2**10
ONE_MB = 2**20
ONE_GB = 2**30

STORED_FILE_NAME_LENGTH = 32
STORED_FILE_NAME_CHARS = 'bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789'

RNG = random.SystemRandom()


@app.template_filter()
def pluralize(s, num):
    # very naive
    if abs(num) == 1:
        return s
    else:
        return s + 's'


@app.template_filter()
def human_size(size):
    if size >= ONE_GB:
        return '{:.1f} GiB'.format(size / ONE_GB)
    elif size >= ONE_MB:
        return '{:.1f} MiB'.format(size / ONE_MB)
    elif size >= ONE_KB:
        return '{:.1f} KiB'.format(size / ONE_KB)
    else:
        return '{} {}'.format(size, pluralize('byte', size))


def gen_unique_id():
    return ''.join(
        RNG.choice(STORED_FILE_NAME_CHARS)
        for _ in range(STORED_FILE_NAME_LENGTH)
    )


# TODO: read these out of the package
ICON_EXTENSIONS = frozenset((
    '7z', 'ai', 'bmp', 'doc', 'docx', 'gif', 'gz', 'html',
    'jpeg', 'jpg', 'midi', 'mp3', 'odf', 'odt', 'pdf', 'png', 'psd', 'rar',
    'rtf', 'svg', 'tar', 'txt', 'wav', 'xls', 'zip', 'unknown',
))


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
