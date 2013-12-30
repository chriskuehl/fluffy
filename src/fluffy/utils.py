from django.conf import settings
from fluffy.backends import FileBackend

backends = {
	"file": FileBackend
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
