#!/usr/bin/env python3
"""Upload files or paste to fluffy.

It can be invoked directly, but is intended to be invoked by two aliases,
"fput" and "fpb". fput uploads files, fpb pastes text.

To install the cli, run `pip install fluffy`.
"""
import argparse
import getpass
import json
import os.path
import re
import sys

import requests
from fluffy_cli import __version__

DESCRIPTION = '''\
fluffy is a simple file-sharing web app. You can upload files, or paste text.

By default, the public instance of fluffy is used: https://fluffy.cc

If you'd like to instead use a different instance (for example, one run
internally by your company), you can specify the --server option.

To make that permanent, you can create a config file with contents similar to:

    {"server": "https://fluffy.my.corp"}

This file can be placed at either /etc/fluffy.json or ~/.config/fluffy.json.
'''


def bold(text):
    if sys.stdout.isatty():
        return '\033[1m{}\033[0m'.format(text)
    else:
        return text


def get_config():
    config = {'server': 'https://fluffy.cc'}
    for path in ('/etc/fluffy.json', os.path.expanduser('~/.config/fluffy.json')):
        try:
            with open(path) as f:
                j = json.load(f)
                if not isinstance(j, dict):
                    raise ValueError(
                        'Expected to parse dict, but the JSON was type "{}" instead.'.format(type(j)),
                    )
                for key, value in j.items():
                    config[key] = value
        except FileNotFoundError:
            pass
        except Exception:
            print(bold('Error parsing config file "{}". Is it valid JSON?'.format(path)))
            raise
    return config


def upload(server, paths, auth):
    files = (('file', sys.stdin.buffer if path == '-' else open(path, 'rb')) for path in paths)
    req = requests.post(
        server + '/upload',
        files=files,
        allow_redirects=False,
        auth=auth,
    )
    assert req.status_code in (301, 302), req.status_code
    print(bold(req.headers['Location']))


def paste(server, path, language, highlight_regex, auth):
    if path == '-':
        content = sys.stdin.read()
    else:
        with open(path) as f:
            content = f.read()

    req = requests.post(
        server + '/paste',
        data={'text': content, 'language': language},
        allow_redirects=False,
        auth=auth,
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
                    else match[0],
                ) for match in squashed
            )

    print(bold(location))


class FluffyArgFormatter(
        argparse.ArgumentDefaultsHelpFormatter,
        argparse.RawDescriptionHelpFormatter,
):
    pass


def upload_main(argv=None):
    config = get_config()
    parser = argparse.ArgumentParser(
        description='Upload files to fluffy.\n\n' + DESCRIPTION,
        formatter_class=FluffyArgFormatter,
    )
    parser.add_argument('--server', default=config['server'], type=str, help='server to upload to')
    parser.add_argument('--version', action='version', version='%(prog)s {}'.format(__version__))
    parser.add_argument('--auth', dest='auth', action='store_true', help='use HTTP Basic auth')
    parser.add_argument('--no-auth', dest='auth', action='store_false', help='do not use HTTP Basic auth')
    parser.set_defaults(auth=config.get('auth', False))
    parser.add_argument(
        '-u', '--username', type=str,
        default=config.get('username', getpass.getuser()),
        help='username for HTTP Basic auth',
    )
    parser.add_argument('file', type=str, nargs='+', help='path to file(s) to upload', default='-')
    args = parser.parse_args(argv)
    auth = None
    if args.auth:
        auth = args.username, getpass.getpass('Password for {}: '.format(args.username))
    return upload(args.server, args.file, auth)


def paste_main(argv=None):
    config = get_config()
    parser = argparse.ArgumentParser(
        description='Paste text to fluffy.\n\n' + DESCRIPTION,
        formatter_class=FluffyArgFormatter,
    )
    parser.add_argument('--server', default=config['server'], type=str, help='server to upload to')
    parser.add_argument('--version', action='version', version='%(prog)s {}'.format(__version__))
    parser.add_argument('-l', '--language', type=str, default='autodetect', help='language for syntax highlighting')
    parser.add_argument('-r', '--regex', type=re.compile, help='regex of lines to highlight')
    parser.add_argument('--auth', dest='auth', action='store_true', help='use HTTP Basic auth')
    parser.add_argument('--no-auth', dest='auth', action='store_false', help='do not use HTTP Basic auth')
    parser.set_defaults(auth=config.get('auth', False))
    parser.add_argument(
        '-u', '--username', type=str,
        default=config.get('username', getpass.getuser()),
        help='username for HTTP Basic auth',
    )
    parser.add_argument('file', type=str, nargs='?', help='path to file to paste', default='-')
    args = parser.parse_args(argv)
    auth = None
    if args.auth:
        auth = args.username, getpass.getpass('Password for {}: '.format(args.username))
    return paste(args.server, args.file, args.language, args.regex, auth)


if __name__ == '__main__':
    if sys.argv[0].endswith('fpb'):
        exit(paste_main())
    else:
        exit(upload_main())
