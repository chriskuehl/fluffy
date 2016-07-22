from setuptools import setup

setup(
    name='fluffy',
    version='1.0.0',
    author='Chris Kuehl',
    author_email='ckuehl@ocf.berkeley.edu',
    py_modules=('fluffy_cli',),
    install_requires={
        'requests',
    },
    classifiers={
        'Programming Language :: Python :: 3',
    },
    entry_points={
        'console_scripts': [
            'fput = fluffy_cli:upload_main',
            'fpb = fluffy_cli:paste_main',
        ],
    },
)
