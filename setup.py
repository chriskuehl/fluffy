from setuptools import find_packages
from setuptools import setup

from fluffy import version


setup(
    name='fluffy-server',
    version=version,
    author='Chris Kuehl',
    author_email='ckuehl@ckuehl.me',
    packages=find_packages(exclude=('test*',)),
    include_package_data=True,
    install_requires=(
        'boto3',
        'cached_property',
        'flask',
        'identify',
        'mistune',
        'pygments',
        'pygments-ansi-color',
        'pyquery',
    ),
    classifiers=(
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6',
    ),
    entry_points={
        'console_scripts': [
            'fluffy-upload-assets = fluffy.component.assets:upload_assets',
        ],
    },
)
