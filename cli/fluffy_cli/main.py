#!/usr/bin/env python3
"""Upload files or paste to fluffy.

It can be invoked directly, but is intended to be invoked by two aliases,
"fput" and "fpb". fput uploads files, fpb pastes text.
"""
import argparse
import re
import sys

import requests
from fluffy_cli import __version__


def bold(text):
    if sys.stdout.isatty():
        return '\033[1m{}\033[0m'.format(text)
    else:
        return text


def upload(server, paths):
    files = (('file', sys.stdin.buffer if path == '-' else open(path, 'rb')) for path in paths)
    req = requests.post(
        server + '/upload',
        files=files,
        allow_redirects=False,
    )
    assert req.status_code in (301, 302), req.status_code
    print(bold(req.headers['Location']))


def paste(server, path, language, highlight_regex):
    if path == '-':
        content = sys.stdin.read()
    else:
        with open(path) as f:
            content = f.read()

    req = requests.post(
        server + '/paste',
        data={'text': content, 'language': language},
        allow_redirects=False,
    )
    assert req.status_code in (301, 302), req.status_code
    location = req.headers['Location']

    if highlight_regex:
        matches = []
        for i, line in enumerate(content.splitlines()):
            if highlight_regex.search(line):
                matches.append(i + 1)

        # squash lines next to each-other
        squashed = []
        for match in matches:
            if not squashed or squashed[-1][1] != match - 1:
                squashed.append([match, match])
            else:
                squashed[-1][1] = match

        if matches:
            location += '#' + ','.join(
                'L{}'.format(
                    '{}-{}'.format(*match)
                    if match[0] != match[1]
                    else match[0]
                ) for match in squashed
            )

    print(bold(location))


def upload_main(argv=None):
    parser = argparse.ArgumentParser(
        description='Upload files to fluffy',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    parser.add_argument('--server', default='https://fluffy.cc', type=str, help='server to upload to')
    parser.add_argument('--version', action='version', version='%(prog)s {}'.format(__version__))
    parser.add_argument('file', type=str, nargs='+', help='path to file(s) to upload', default='-')
    args = parser.parse_args(argv)
    return upload(args.server, args.file)


def paste_main(argv=None):
    parser = argparse.ArgumentParser(
        description='Paste text to fluffy',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )
    parser.add_argument('--server', default='https://fluffy.cc', type=str, help='server to upload to')
    parser.add_argument('--version', action='version', version='%(prog)s {}'.format(__version__))
    parser.add_argument('-l', '--language', type=str, default='autodetect', help='language for syntax highlighting')
    parser.add_argument('-r', '--regex', type=re.compile, help='regex of lines to highlight')
    parser.add_argument('file', type=str, nargs='?', help='path to file to paste', default='-')
    args = parser.parse_args(argv)
    return paste(args.server, args.file, args.language, args.regex)


if __name__ == '__main__':
    if sys.argv[0].endswith('fpb'):
        exit(paste_main())
    else:
        exit(upload_main())
