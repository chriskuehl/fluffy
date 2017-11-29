from fluffy_cli import __version__
from setuptools import find_packages
from setuptools import setup


def main():
    setup(
        name='fluffy',
        version=__version__,
        author='Chris Kuehl',
        author_email='ckuehl@ckuehl.me',
        packages=find_packages(exclude=('test*',)),
        install_requires=(
            'requests',
        ),
        classifiers=(
            'Programming Language :: Python :: 3',
        ),
        entry_points={
            'console_scripts': [
                'fput = fluffy_cli.main:upload_main',
                'fpb = fluffy_cli.main:paste_main',
            ],
        },
    )


if __name__ == '__main__':
    exit(main())
