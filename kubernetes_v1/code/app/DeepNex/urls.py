"""DeepNex URL Configuration

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/1.10/topics/http/urls/
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
from django.conf.urls import url
from django.contrib import admin
from django.contrib.auth import views as auth_views

from controller import views

urlpatterns = [
    url(r'^login/$', auth_views.login, name='login'),
    url(r'^logout/$', auth_views.logout_then_login, name='logout'),
    url(r'^admin/', admin.site.urls),
    url(r'^$', views.home),
    url(r'^home/$', views.home, name='home'),
    url(r'^workspace/$', views.workspace),
    url(r'^apps/$', views.apps),
    url(r'^app_manage$', views.app_manage),
    url(r'^app_info$', views.app_info),
    url(r'^profile/$', views.profile),
    url(r'^user_profile$', views.user_profile),
    url(r'^usage/$', views.usage),
    url(r'^images/$', views.images),
    url(r'^image_manage$', views.image_manage),
    url(r'^grafana_data/(?P<url>.*)$', views.grafana_data),
    url(r'^dashboard_data/(?P<url>.*)$', views.dashboard_data),
    url(r'^app_data/(?P<url>.*)$', views.api_appdata),
    url(r'^app_data_allothers/(?P<url>.*)$', views.api_appdata_allothers)
]
