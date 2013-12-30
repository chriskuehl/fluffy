import os

class FileBackend:
	"""FileBackend is a storage backend which stores files and info pages
	on the local disk."""

	def __init__(self, options):
		self.options = options
		self.validate()

	def validate(self):
		"""Validate the options given to the backend, raising an exception
		if anything isn't right."""

		for path_name in ["file_path", "info_path"]:
			if path_name not in self.options:
				raise Exception("No {} given.".format(path_name))

			path = self.options[path_name]

			if path.endswith("/"):
				path = path[:-1]
				self.options[path_name] = path

			if not os.path.isdir(path):
				raise Exception("Invalid {}: {}".format(path_name, path))

	def store(self, stored_file):
		"""Stores the file and its info page. This is the only method
		which needs to be called in order to persist the uploaded file to
		the storage backend."""
		path = self.file_path(stored_file)
		print("Writing to {}...".format(path))

		with open(path, "wb+") as dest:
			for chunk in stored_file.file.chunks():
				dest.write(chunk)

	def file_path(self, stored_file):
		return "{}/{}".format(self.options["file_path"], stored_file.name)
