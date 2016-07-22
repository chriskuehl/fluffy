from setuptools import find_packages
from setuptools import setup

setup(
    name='fluffy-server',
    version='1.2.1',
    author='Chris Kuehl',
    author_email='ckuehl@ocf.berkeley.edu',
    packages=find_packages(),
    include_package_data=True,
    install_requires={
        'boto3',
        'cached_property',
        'flask',
        'pygments',
    },
    classifiers={
        'Programming Language :: Python :: 3',
    },
    entry_points={
        'console_scripts': [
            'fluffy-upload-assets = fluffy.assets:upload_assets',
        ],
    },
)
