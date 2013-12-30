import os

"""Backends handle storing of uploaded files. A backend should implement
__init__(self, options) where options will be the dict of options given in the
fluffy settings, and store(self, stored_file) which stores an uploaded file.

All other details are left up to your implementation."""

class FileBackend:
	"""FileBackend is a storage backend which stores files and info pages
	on the local disk."""

	def __init__(self, options):
		self.options = options

	def store(self, stored_file):
		"""Stores the file and its info page. This is the only method
		which needs to be called in order to persist the uploaded file to
		the storage backend."""
		path = self.options["file_path"].format(name=stored_file.name)

		# store the file itself
		print("Writing to {}...".format(path))
		with open(path, "wb+") as dest:
			for chunk in stored_file.file.chunks():
				dest.write(chunk)
