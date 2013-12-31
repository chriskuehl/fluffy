## fluffy: a file sharing web app that doesn't suck.

![fluffy screenshots](http://i.fluffy.cc/sx8c22NDDBw2hG0slzZVLM2ZW2FHw0j5.png)

### What is fluffy?

fluffy is a Django-based web application that allows you to upload arbitrary
files to the web. Once you upload the files, you get a link to the file which
you can share.

The reference instance of fluffy is online at [fluffy.cc](http://fluffy.cc/).
You can also run your own!

### What isn't fluffy?

* **fluffy isn't social.** Files are given a long, random (unguessable) name.
  There's no upload feed or list of files.
* **fluffy isn't intrusive.** Your files aren't resized, compressed, stripped,
  or modified in any way.
* **fluffy isn't annoying.** A simple, modern page for uploading your files. No
  ads, no memes, and no animated cat GIFs.

### Technical philosophy

fluffy only handles uploading and storing your files. There's no database, and
it's up to you to figure out how you serve the uploaded files. Once fluffy
stores a file, it forgets about it.

This not only makes the code simple, but also makes maintenance easy. If you
wish to stop accepting uploads, you can easily throw the existing uploads on
any web server to ensure their continued availability.

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
single method. The current backends are both less than 30 lines of code.

#### Serving files

fluffy won't serve your files, period. It's up to you to figure this part out.
Depending on which backend you use, you may get it easily. For example, Amazon
S3 makes it easy to serve uploaded files via the web.

### Run your own fluffy

To host your own copy of fluffy, clone the git repo. Copy `settings.py.tmpl` to
`settings.py` and adjust it to your needs, being sure to uncomment whichever
storage backend you wish to use. There's no database, so setup is very simple.

Once you've adjusted the configuration, you can deploy fluffy the way you
deploy any Django app. fluffy is tested with Python versions 2.7 and 3.3, and
is also probably compatible with other versions.

Since "fluffy.cc" is just one instance of fluffy, you'll probably want to
change the page headers to feature your own site.

### Contributing, license, and credits

Contributions to fluffy are welcome! Send your pull requests or file an issue.
Thanks for the help!

fluffy is copyright &copy; 2013 Chris Kuehl, with all original code released
under an MIT license. Other code, such as included libraries or Django starter
code, is under its original license. See LICENSE for full details.

fluffy uses awesome icon sets developed by
[FatCow](http://www.fatcow.com/free-icons) and [Everaldo
Coelho](http://www.everaldo.com/).
