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

Additionally, Debian packages for the command-line tools are available in the
[GitHub releases tab](https://github.com/chriskuehl/fluffy/releases). These
packages contain no binary components and should be compatible with most
releases of Debian and Ubuntu.


### Contributing, license, and credits

Contributions to fluffy are welcome! Send your pull requests or file an issue.
Thanks for the help!

fluffy is released under the MIT license; see LICENSE for full details.


#### Running locally for development

To run fluffy during development, run `make venv` and then `pgctl start`.
You should now have fluffy running at `http://localhost:5000`.


### FAQ
#### Why are there only certain languages in the dropdown? Can I add more?

Since it's just a normal `<option>` dropdown now, I didn't want to have all
hundreds of languages that Pygments supports, as I thought that would make the
UI worse for little benefit. Instead, currently there's a hand-picked list of
languages that I thought were most popular (but it's definitely biased toward
what I use!).

In the long term, I'd love to replace the dropdown with something smarter
(maybe a JS dropdown with all the possible languages, featuring the most
popular at the top, but with all available below, or with autocomplete or
something).

In the medium term, definitely feel free to open an issue or send a PR to add
another language. I'll happily merge it.

As a workaround, note that the "automatically detect" can detect languages not
in the dropdown (but it's not very accurate much of the time, unfortunately).
Additionally, if you use the CLI, you can pass `-l <language>` and use any
language supported by Pygments.


#### Why are there only a few themes to choose from in the pastebin?

Mostly it's just lack of time to add more. If you have a Pygments theme you
like, please open an issue or PR, I'll definitely help get it added.

Primarily the reasons are:

* There are a lot of Pygments themes and many of them are (imo) low-quality or
  extremely similar. I thought it was better to hand-curate them and have a
  small-ish number, rather than a large number where it's hard to find the
  really good themes.

* It's not quite as easy as just adding the style name to a list. Fluffy also
  needs to know colors to use for borders, line numbers, highlights, diff
  highlights, hover versions of all of these, ANSI color codes that work, etc.
  There's something like 30 constants to set per theme.

* With the current implementation, every new theme bloats the CSS a little
  more. (There's probably a technical solution to this.)


fluffy uses awesome icon sets developed by
[FatCow](http://www.fatcow.com/free-icons) and
[Everaldo Coelho](http://www.everaldo.com/).
