from django.conf import settings

def constants(request):
	return {
		"abuse_contact": settings.ABUSE_CONTACT,
		"home_url": settings.HOME_URL
	}
