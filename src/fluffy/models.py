import random

class StoredFile:
	"""A File object wraps an actual file and has a unique ID."""
	NAME_LENGTH = 32
	NAME_CHARS = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"

	def __init__(self, file):
		self.generate_name()

	def generate_name(self):
		"""Generates a unique name for the file. We don't actually verify that
		the name is unique, but chances are very slim that it won't be."""
		self.name = "".join(random.choice(StoredFile.NAME_CHARS) \
			for _ in range(StoredFile.NAME_LENGTH))
