import json
import zlib
import base64
from django.conf import settings
from fluffy.backends import FileBackend, S3CommandLineBackend

backends = {
	"file": FileBackend,
	"s3cli": S3CommandLineBackend
}

def get_backend():
	"""Returns a backend instance as configured in the settings."""
	conf = settings.STORAGE_BACKEND
	name, options = conf["name"], conf["options"]

	return backends[name](options)


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

def encode_obj(obj):
	"""Encodes an object for inclusion in a query string.

	Works internally by serializing the object into JSON, gzipping the JSON,
	and then base64 encoding it.
	"""
	serialized = json.dumps(obj)
	zipped = zlib.compress(serialized.encode("utf-8"))

	return base64.urlsafe_b64encode(zipped).decode("utf-8")

def decode_obj(encoded):
	"""Decodes a string encoded by encode_obj into an object."""
	zipped = base64.urlsafe_b64decode(encoded.encode("utf-8"))
	serialized = zlib.decompress(zipped).decode("utf-8")

	return json.loads(serialized)
