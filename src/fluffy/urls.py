from django.conf.urls import patterns, include, url

urlpatterns = patterns('',
    # Examples:
    # url(r'^$', 'fluffy.views.home', name='home'),
    # url(r'^blog/', include('blog.urls')),
	url(r'^$', 'fluffy.views.index', name='index'),
	url(r'^upload$', 'fluffy.views.upload', name='upload')
)
