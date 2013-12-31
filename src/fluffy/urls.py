from django.conf.urls import patterns, include, url

urlpatterns = patterns('',
	url(r'^$', 'fluffy.views.index', name='index'),
	url(r'^upload$', 'fluffy.views.upload', name='upload'),
	url(r'^details/(?P<enc>.*)$', 'fluffy.views.details', name='details')
)
