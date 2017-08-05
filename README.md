## fluffy: a file sharing web app that doesn't suck.

| Overall build status | [![Build Status](https://travis-ci.org/chriskuehl/fluffy.svg?branch=master)](https://travis-ci.org/chriskuehl/fluffy) [![Coverage Status](https://coveralls.io/repos/github/chriskuehl/fluffy/badge.svg?branch=master)](https://coveralls.io/github/chriskuehl/fluffy?branch=master) |
| -------------------- | ----------------------------- |
| `fluffy-server`      | [![PyPI version](https://badge.fury.io/py/fluffy-server.svg)](https://pypi.python.org/pypi/fluffy-server) |
| `fluffy`             | [![PyPI version](https://badge.fury.io/py/fluffy.svg)](https://pypi.python.org/pypi/fluffy) |


![fluffy screenshots](https://i.fluffy.cc/sx8c22NDDBw2hG0slzZVLM2ZW2FHw0j5.png)


### What is fluffy?

fluffy is a Flask-based web application that allows you to upload arbitrary
files to the web. Once you upload the files, you get a link to the file which
you can share.

The reference instance of fluffy is online at [fluffy.cc](https://fluffy.cc/).
You can also run your own!


### What isn't fluffy?

* **fluffy isn't social.** Files are given a long, random (unguessable) name.
  There's no upload feed or list of files.
* **fluffy isn't intrusive.** Your files aren't resized, compressed, stripped,
  or modified in any way.
* **fluffy isn't annoying.** A simple, modern page for uploading your files. No
  ads, no memes, and no comments.


### Philosophy and motivation

fluffy was created out of frustration from seeing hundreds of files (mostly
images) be lost or deleted over the years from popular image hosts such as
imageshack which either deleted files or closed their doors entirely. Fluffy is
designed so that it is easy to stop accepting uploads while still serving
existing files, with the hope being that a "shut down" would involve no longer
accepting uploads, but still continuing to serve existing uploads.

fluffy only handles uploading and storing your files. There's no database, and
it's up to you to figure out how you serve the uploaded files. Once fluffy
stores a file, it forgets about it.

This not only makes the code simple, but also makes maintenance easy. If you
wish to stop accepting uploads, you can easily throw the existing uploads on S3
or any web server to ensure their continued availability.

This does make some features hard or impossible to implement, however, so if
you want to do anything post-upload at the application level, fluffy probably
isn't for you.


#### Storing files

fluffy hands off uploaded files to a *storage backend*, which is responsible
for saving the file. The following backends are currently available:

* **File.** Storage on the local filesystem. You can easily serve these with
  any web server.
* **Amazon S3.** Storage on Amazon S3. You can serve these with S3 static
  websites or with CloudFront (if you want a CDN).

Writing a storage backend is dead simple and requires you to implement only a
single method. The current backends are both about ten lines of code.


#### Serving files

fluffy won't serve your files, period. It's up to you to figure this part out.
Depending on which backend you use, you may get it easily. For example, Amazon
S3 makes it easy to serve uploaded files via the web.


### Run your own fluffy

There's a public "reference implementation" of fluffy at
[fluffy.cc](https://fluffy.cc/).

To host your own copy of fluffy, just adjust `settings.py` to your needs, being
sure to uncomment whichever storage backend you wish to use. There's no
database, so setup is very simple.

Once you've adjusted the configuration, you can deploy fluffy the way you
deploy any Flask app. fluffy is tested with Python versions 3.5 and 3.6.


### Command-line uploading tools

Two tools, `fput` and `fpb`, are provided. They can be installed with `pip
install fluffy` and used from the command line. Use `--help` with either tool
for more information.


### Contributing, license, and credits

Contributions to fluffy are welcome! Send your pull requests or file an issue.
Thanks for the help!

fluffy is released under the MIT license; see LICENSE for full details.


#### Running locally for development

To run fluffy during development, run `make venv` and then `pgctl start`.
You should now have fluffy running at `http://localhost:5000`.


fluffy uses awesome icon sets developed by
[FatCow](http://www.fatcow.com/free-icons) and
[Everaldo Coelho](http://www.everaldo.com/).
