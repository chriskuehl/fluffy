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

			if not os.path.isdir(path):
				raise Exception("Invalid {}: {}".format(path_name, path))

	def store(self, file):
		pass
