from django.shortcuts import render
from django.http import HttpResponse
from fluffy.models import StoredFile

def index(request):
	return render(request, "index.html")

def upload(request):
	files = []

	for file in request.FILES.getlist("file"):
		sfile = StoredFile(file)
		files.append(sfile)

	return HttpResponse("test")
