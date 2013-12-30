import random
import os

class StoredFile:
	"""A File object wraps an actual file and has a unique ID."""
	NAME_LENGTH = 32
	NAME_CHARS = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"

	def __init__(self, file):
		self.file = file
		self.generate_name()

	def generate_name(self):
		"""Generates a unique name for the file. We don't actually verify that
		the name is unique, but chances are very slim that it won't be."""
		name = "".join(random.choice(StoredFile.NAME_CHARS) \
			for _ in range(StoredFile.NAME_LENGTH))

		extension = self.extension

		if extension:
			name += "." + extension

		self.name = name

	@property
	def extension(self):
		"""Returns extension without leading period, or empty string if no
		extension."""
		path, ext = os.path.splitext(self.file.name)
		return ext[1:] if ext else ""
