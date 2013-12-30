from django.conf import settings

def abuse_contact(request):
	return {"abuse_contact": settings.ABUSE_CONTACT}
