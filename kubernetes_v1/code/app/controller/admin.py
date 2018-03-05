from django.contrib import admin
from django.contrib.auth.models import Group, User
from .models import Docker, Profile, CustomUserAdmin, Server

# Register your models here.

admin.site.register(Profile)
admin.site.unregister(Group)
admin.site.unregister(User)
admin.site.register(User, CustomUserAdmin)
admin.site.register(Server)
