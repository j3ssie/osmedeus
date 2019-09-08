"""rest URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/1.11/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  url(r'^$', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  url(r'^$', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.conf.urls import url, include
    2. Add a URL to urlpatterns:  url(r'^blog/', include('blog.urls'))
"""
from django.contrib import admin
from django.urls import path
from django.conf.urls import url, include
# from django.contrib.staticfiles.urls import staticfiles_urlpatterns
from django.conf.urls import handler404, handler500


from rest_framework_simplejwt.views import (
    TokenObtainPairView,
    TokenRefreshView,
)

# from rest.views.view import testing
from rest.frontend import index, error_404, error_500

urlpatterns = [
    url(r'^django-admin/', admin.site.urls),

    # auth url
    url(r'^auth/api/token/$', TokenObtainPairView.as_view(),
        name='token_obtain_pair'),
    url(r'^auth/api/token/refresh/$',
        TokenRefreshView.as_view(), name='token_refresh'),

    # core endpoints
    path('', include('api.urls')),
    # path('', include('action.urls')),

    # serve react
    url(r'^.*', index, name="index"),
]

handler404 = 'rest.frontend.error_404'
handler500 = 'rest.frontend.error_500'
