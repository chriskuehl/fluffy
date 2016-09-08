from setuptools import find_packages
from setuptools import setup

from fluffy import version


setup(
    name='fluffy-server',
    version=version,
    author='Chris Kuehl',
    author_email='ckuehl@ocf.berkeley.edu',
    packages=find_packages(),
    include_package_data=True,
    install_requires={
        'boto3',
        'cached_property',
        'flask',
        'mistune',
        'pygments',
        'pyquery',
    },
    classifiers={
        'Programming Language :: Python :: 3',
    },
    entry_points={
        'console_scripts': [
            'fluffy-upload-assets = fluffy.component.assets:upload_assets',
        ],
    },
)
