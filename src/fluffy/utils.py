import os
import json
import zlib
import base64
from django.conf import settings

ONE_GB = 1073741824
ONE_MB = 1048576
ONE_KB = 1024

def get_human_size(size):
	"""Convert a byte count into a human-readable size string like "4.2 MB".

	Roughly based on Apache Commons FileUtils#byteCountToDisplaySize:
	https://commons.apache.org/proper/commons-io/ """

	if size / ONE_GB >= 1:
		return "{:.1f} GB".format(size / ONE_GB)
	elif size / ONE_MB >= 1:
		return "{:.1f} MB".format(size / ONE_MB)
	elif size / ONE_KB >= 1:
		return "{:.1f} KB".format(size / ONE_KB)
	else:
		return "{} bytes".format(size)

# constants to represent zipped or unzipped (at start of encoded string)
NOT_ZIPPED = "u"
ZIPPED = "z"

def encode_obj(obj):
	"""Encodes an object for inclusion in a query string.

	Works internally by serializing the object into JSON, gzipping the JSON,
	and then base64 encoding it.

	The serialized data will only be zipped if doing so would result in a
	shorter string. The returned string will start with ZIPPED or NOT_ZIPPED
	(a single character constant) to indicate zipped or unzipped.
	"""
	serialized = json.dumps(obj)
	zipped = zlib.compress(serialized.encode("utf-8"))

	# pick either zipped or unzipped, whichever is shortest
	encode = lambda s: base64.urlsafe_b64encode(s).decode("utf-8")

	choices = (
		NOT_ZIPPED + encode(serialized.encode("utf-8")),
		ZIPPED + encode(zipped)
	)

	return min(choices, key=len)

def decode_obj(encoded):
	"""Decodes a string encoded by encode_obj into an object."""
	zipped = encoded.startswith(ZIPPED)
	encoded = encoded[1:]

	bytes = base64.urlsafe_b64decode(encoded.encode("utf-8"))

	if zipped:
		text = zlib.decompress(bytes).decode("utf-8")
	else:
		text = bytes.decode("utf-8")

	return json.loads(text)


ICON_EXTENSIONS = [
	"7z", "ai", "bmp", "doc", "docx", "gif", "gz", "html",
	"jpeg", "jpg", "midi", "mp3", "odf", "odt", "pdf", "png", "psd", "rar",
	"rtf", "svg", "tar", "txt", "wav", "xls", "zip"
]

def get_extension_icon(extension):
	"""Returns the filename of the icon to use for an extension, excluding
	the .png at the end of the icon name.
	"""
	extension = extension.lower()
	return extension if extension in ICON_EXTENSIONS else "unknown"

def trim_filename(name, length):
	"""Trim a filename down to a desired maximum length, making attempts to
	preserve the important parts of the name.

	We prefer to preserve, in order:
	1. the extension in full
	2. the first three and last two characters of the name
	3. everything else, prioritizing the first characters
	"""
	if len(name) <= length:
		return name

	name, ext = os.path.splitext(name)
	length = max(length, len(ext) + 5)

	if len(name) + len(ext) <= length + 2:
		# we can't really do better
		return name + ext

	length += 1 # don't count the dot in the extension
	length += 1 # count elipses as two characters

	prefix = ""
	suffix = name[-2:]

	get_result = lambda: prefix.strip() + "..." + suffix.strip() + ext

	for i in range(3, len(name) - 2):
		prefix = name[:i]
		result = get_result()

		if len(result) >= length:
			return result

	return get_result()

def validate_files(file_list):
	for file in file_list:
		validate_file(file)

def validate_file(file):
	if file.size > settings.MAX_UPLOAD_SIZE:
		human_size = get_human_size(settings.MAX_UPLOAD_SIZE)
		msg = "{} exceeded the maximum file size limit of {}"
		msg = msg.format(file.name, human_size)

		raise ValidationException(msg)

class ValidationException(Exception):
	pass
