# Django, uWSGI and Nginx in a container, using Supervisord

This Dockerfile is modified for DeepNex deployment.

Update all code in submodules

    git submodule init # Only need to run one time, under all folders that have submodules
    git pull --recurse-submodules  && git submodule update --recursive

### Preparation
Before the build process. Please generate the static files for system.

    python manage.py collectstatic # Run under the app folder

### Build and run
* docker build -t yaoman3/deepnex .
* docker run -v {NFS path}:/data -v {config file}:/home/docker/code/app/k8s_api/config.py -p 80:80 -d --restart=always yaoman3/deepnex

### GPU support
Need to setup the /var/deepnex/gpu file in every k8s server.

### How to insert your application

The django server is in /app.

