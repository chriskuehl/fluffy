from setuptools import find_packages
from setuptools import setup

from fluffy import version


with open('requirements-minimal.txt') as f:
    minimal_reqs = f.read().splitlines()


setup(
    name='fluffy-server',
    version=version,
    author='Chris Kuehl',
    author_email='ckuehl@ckuehl.me',
    packages=find_packages(exclude=('test*', 'playground')),
    include_package_data=True,
    install_requires=minimal_reqs,
    classifiers=(
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.9',
    ),
    entry_points={
        'console_scripts': [
            'fluffy-upload-assets = fluffy.component.assets:upload_assets',
        ],
    },
)
