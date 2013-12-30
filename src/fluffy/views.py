import time
from django.shortcuts import render
from django.http import HttpResponse
from fluffy.models import StoredFile
from fluffy.utils import get_backend

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

	return HttpResponse("test")
