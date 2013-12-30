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
