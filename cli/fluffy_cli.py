#!/usr/bin/env python3
"""Upload files or paste to fluffy.

It can be invoked directly, but is intended to be invoked by two aliases,
"fput" and "fpb". fput uploads files, fpb pastes text.
"""
import argparse
import sys

import requests


def bold(text):
    if sys.stdout.isatty():
        return '\033[1m{}\033[0m'.format(text)
    else:
        return text


def upload(server, paths):
    files = (('file', sys.stdin if path == '-' else open(path, 'rb')) for path in paths)
    req = requests.post(
        server + '/upload',
        files=files,
        allow_redirects=False,
    )
    assert req.status_code in (301, 302), req.status_code
    print(bold(req.headers['Location']))


def paste(server, path, language):
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
    print(bold(req.headers['Location']))


def upload_main(argv=None):
    parser = argparse.ArgumentParser(description='Upload files to fluffy')
    parser.add_argument('--server', default='http://fluffy.cc', type=str, help='server to upload to')
    parser.add_argument('file', type=str, nargs='+', help='path to file(s) to upload', default='-')
    args = parser.parse_args(argv)
    return upload(args.server, args.file)


def paste_main(argv=None):
    parser = argparse.ArgumentParser(description='Paste text to fluffy')
    parser.add_argument('--server', default='http://fluffy.cc', type=str, help='server to upload to')
    parser.add_argument('-l', '--language', type=str, default='autodetect')
    parser.add_argument('file', type=str, nargs='?', help='path to file to paste', default='-')
    args = parser.parse_args(argv)
    return paste(args.server, args.file, args.language)


if __name__ == '__main__':
    if sys.argv[0].endswith('fpb'):
        exit(paste_main())
    else:
        exit(upload_main())
