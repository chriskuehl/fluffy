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
    packages=find_packages(exclude=('test*',)),
    include_package_data=True,
    install_requires=minimal_reqs,
    license='Apache License 2.0',
    classifiers=(
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.10',
        'Programming Language :: Python :: 3.11',
    ),
    python_requires='>=3.10',
    entry_points={
        'console_scripts': [
            'fluffy-upload-assets = fluffy.component.assets:upload_assets',
        ],
    },
)
