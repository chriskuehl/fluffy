import os
import pipes

"""Backends handle storing of uploaded files. A backend should implement
__init__(self, options) where options will be the dict of options given in the
fluffy settings, and store(self, stored_file) which stores an uploaded file.

All other details are left up to your implementation."""

class FileBackend:
	"""Storage backend which stores files and info pages on the local disk."""

	def __init__(self, options):
		self.options = options

	def store(self, stored_file):
		"""Stores the file and its info page. This is the only method
		which needs to be called in order to persist the uploaded file to
		the storage backend."""
		path = self.options["file_path"].format(name=stored_file.name)
		info_path = self.options["info_path"].format(name=stored_file.name)

		# store the file itself
		print("Writing to {}...".format(path))
		with open(path, "wb+") as dest:
			for chunk in stored_file.file.chunks():
				dest.write(chunk)

		# store the info page
		print("Writing info page to {}...".format(info_path))
		with open(info_path, "wb+") as dest:
			dest.write(stored_file.info_html.encode("utf-8"))

class S3CommandLineBackend:
	"""Storage backend which uploads to S3 using AWS' command-line tools.

	We use the command-line tools because at the time of writing, boto does not
	support python3. Once this changes, it will be trivial to switch out the
	commands used for uploading.

	For this backend to work, you must have awscli installed and configured.

	For installation, try: pip install awscli
	For configuration, try: aws configure

	To verify everything works, try: aws s3 ls
	You should see a list of your S3 buckets."""

	def __init__(self, options):
		self.options = options

	def store(self, stored_file):
		"""Stores the file and its info page. This is the only method
		which needs to be called in order to persist the uploaded file to
		the storage backend."""

		def write_file(dest):
			for chunk in stored_file.file.chunks():
				dest.write(chunk)

		def write_info(dest):
			dest.write(stored_file.info_html.encode("utf-8"))

		files = (
			{"name": "file", "write": write_file},
			{"name": "info", "write": write_info}
		)

		for file in files:
			name = self.options[file["name"] + "_name"].format(name=stored_file.name)
			path = self.options["tmp_path"].format(name=name)

			print("Writing temp file '{}' to '{}'".format(name, path))

			with open(path, "wb+") as dest:
				file["write"](dest)

			s3 = self.options[file["name"] + "_s3path"].format(name=stored_file.name)

			cmd = "aws s3 cp {} {}".format(pipes.quote(path), pipes.quote(s3))
			print("Uploading to S3 with command: {}".format(cmd))
			status = os.system(cmd)
			print("Status: {}".format(status))

			os.remove(path)
