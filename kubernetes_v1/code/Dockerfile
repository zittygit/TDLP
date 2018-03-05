# Copyright 2017 Yaoman Li
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM ubuntu:14.04

MAINTAINER Dockerfiles

# Install required packages and remove the apt packages cache when done.

RUN apt-get update && apt-get install -y \
	git \
	python3 \
	python3-dev \
	python3-setuptools \
	nginx \
	supervisor \
	sqlite3 \
  && rm -rf /var/lib/apt/lists/*

RUN apt-get update --fix-missing && apt-get install -y wget bzip2 ca-certificates \
    libglib2.0-0 libxext6 libsm6 libxrender1

RUN echo 'export PATH=/opt/conda/bin:$PATH' > /etc/profile.d/conda.sh && \
    wget --quiet https://repo.continuum.io/archive/Anaconda3-4.3.1-Linux-x86_64.sh -O ~/anaconda.sh && \
    /bin/bash ~/anaconda.sh -b -p /opt/conda && \
    rm ~/anaconda.sh

ENV PATH /opt/conda/bin:$PATH

# RUN easy_install pip

# install uwsgi now because it takes a little while
RUN pip install uwsgi

# setup all the configfiles
RUN echo "daemon off;" >> /etc/nginx/nginx.conf
COPY nginx-app.conf /etc/nginx/sites-available/default
COPY supervisor-app.conf /etc/supervisor/conf.d/

# COPY requirements.txt and RUN pip install BEFORE adding the rest of your code, this will cause Docker's caching mechanism
# to prevent re-installinig (all your) dependencies when you made a change a line or two in your app. 

COPY app/requirements.txt /home/docker/code/app/
RUN pip install -r /home/docker/code/app/requirements.txt

# add (the rest of) our code
COPY . /home/docker/code/

WORKDIR /home/docker/code/app/

RUN python manage.py collectstatic && python -m compileall -b . && find */* -name '*.py' -delete

WORKDIR /

# install django, normally you would remove this step because your project would already
# be installed in the code/app/ directory
# RUN django-admin.py startproject website /home/docker/code/app/ 


EXPOSE 80
CMD ["supervisord", "-n"]
