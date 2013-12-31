import random
import os
from fluffy.utils import get_human_size, trim_filename, get_extension_icon
from datetime import date
from django.template.loader import render_to_string
from django.conf import settings

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
	def info_html(self):
		"""Returns the HTML of the info page."""
		extension = self.extension

		params = {
			"name": trim_filename(self.file.name, 17),
			"size": get_human_size(self.file.size),
			"date": date.today().strftime("%B %e, %Y"),
			"extension": get_extension_icon(extension),
			"download_url": settings.FILE_URL.format(name=self.name)
		}

		return render_to_string("info.html", params)

	@property
	def extension(self):
		"""Returns extension without leading period, or empty string if no
		extension."""

		ext = os.path.splitext(self.file.name)[1]
		return ext[1:] if ext else ""
