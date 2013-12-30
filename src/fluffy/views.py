import time
import os
from django.shortcuts import render
from django.http import HttpResponse
from fluffy.models import StoredFile
from fluffy.utils import get_backend, get_human_size, encode_obj

def index(request):
	return render(request, "index.html")

def upload(request):
	"""Process an upload, storing each uploaded file with the configured
	storage backend. Redirects to a status page which displays the uploaded
	file(s).

	Returns JSON containing the URL to redirect to; the actual redirect is done
	by the JavaScript on the upload page.
	"""
	backend = get_backend()
	file_list = request.FILES.getlist("file")
	stored_files = [StoredFile(file) for file in file_list]

	start = time.time()
	print("Storing {} files...".format(len(stored_files)))

	for stored_file in stored_files:
		print("Storing {}...".format(stored_file.name))
		backend.store(stored_file)

	elapsed = time.time() - start
	print("Stored {} files in {:.1f} seconds.".format(len(stored_files), elapsed))

	response_list = [get_response(f) for f in stored_files]
	response = encode_obj(response_list)

	return HttpResponse(response)

def get_response(stored_file):
	"""Returns a tuple of details of a single stored file to be included in the
	parameters of the info page.

	Details in the tuple:
	  - stored name
	  - human name without extension (to save space)
	  - human size (to save space)
	"""
	human_name = os.path.splitext(stored_file.file.name)[0]
	human_size = get_human_size(stored_file.file.size)

	return (stored_file.name, human_name, human_size)
