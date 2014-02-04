import time
import os
import json
from django.core.urlresolvers import reverse
from django.shortcuts import render, redirect
from django.http import HttpResponse
from django.conf import settings
from fluffy.models import StoredFile
from fluffy.utils import get_human_size, encode_obj, decode_obj, \
	trim_filename, get_extension_icon, validate_files, ValidationException
from fluffy.backends import get_backend, BackendException

def index(request):
	return render(request, "index.html")

def upload(request):
	"""Process an upload, storing each uploaded file with the configured
	storage backend. Redirects to a status page which displays the uploaded
	file(s).

	Returns JSON containing the URL to redirect to; the actual redirect is done
	by the JavaScript on the upload page.
	"""
	try:
		backend = get_backend()
		file_list = request.FILES.getlist("file")

		validate_files(file_list);

		stored_files = [StoredFile(file) for file in file_list]

		start = time.time()
		print("Storing {} files...".format(len(stored_files)))

		for stored_file in stored_files:
			print("Storing {}...".format(stored_file.name))
			backend.store(stored_file)

		elapsed = time.time() - start
		print("Stored {} files in {:.1f} seconds.".format(len(stored_files), elapsed))

		# redirect to the details page if multiple files were uploaded
		# otherwise, redirect to the info page of the single file
		if len(stored_files) > 1:
			details = [get_details(f) for f in stored_files]
			details_encoded = encode_obj(details)

			url = reverse("details", kwargs={"enc": details_encoded})
		else:
			url = settings.INFO_URL.format(name=stored_files[0].name)


		response = {
			"success": True,
			"redirect": url
		}
	except BackendException as e:
		print("Error storing files: {}".format(e))
		print("\t{}".format(e.display_message))

		response = {
			"success": False,
			"error": e.display_message
		}
	except ValidationException as e:
		print("Refusing to accept file (failed validation):")
		print("\t{}".format(e))

		response = {
			"success": False,
			"error": str(e)
		}
	except Exception as e:
		print("Unknown error storing files: {}".format(e))

		response = {
			"success": False,
			"error": "An unknown error occured."
		}

	if "json" in request.GET:
		return HttpResponse(json.dumps(response), content_type="application/json")
	else:
		if not response["success"]:
			return HttpResponse("Error: {}".format(response["error"]), content_type="text/plain")
		else:
			return redirect(response["redirect"])

def get_details(stored_file):
	"""Returns a tuple of details of a single stored file to be included in the
	parameters of the info page.

	Details in the tuple:
	  - stored name
	  - human name without extension (to save space)
	"""
	human_name = os.path.splitext(stored_file.file.name)[0]

	return (stored_file.name, human_name)

def details(request, enc=encode_obj([])):
	"""Displays details about an upload (or any set of files, really).

	enc is the encoded list of detail tuples, as returned by get_details.
	"""
	req_details = decode_obj(enc)
	details = [get_full_details(file) for file in req_details]

	return render(request, "details.html", {"details": details})

def get_full_details(file):
	"""Returns a dictionary of details for a file given a detail tuple."""
	stored_name = file[0]
	name = file[1] # original file name
	ext = os.path.splitext(stored_name)[1]

	return {
		"download_url": settings.FILE_URL.format(name=stored_name),
		"info_url": settings.INFO_URL.format(name=stored_name),
		"name": trim_filename(name + ext, 17), # original name is stored w/o extension
		"extension": get_extension_icon(ext[1:] if ext else "")
	}
