from django.shortcuts import render
from django.shortcuts import redirect
from django.conf import settings


# serve react app
def index(request):
    return render(request, 'index.html')


def error_404(request, exceptions):
    return render(request, "error/404.html", status=404)


def error_500(request):
    return render(request, "error/500.html", status=500)
