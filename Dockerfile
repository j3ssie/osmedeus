FROM debian:latest

RUN apt-get update && \
    apt-get -qq install \ 
    npm \
    locales \
    git \
    sudo \
    wget \
    python3-pip \
    python-pip \
    curl \
    libcurl4-openssl-dev \
    bsdmainutils \
    xsltproc \
    build-essential

# Set the locale
RUN sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    dpkg-reconfigure --frontend=noninteractive locales && \
    update-locale LANG=en_US.UTF-8

ENV LANG en_US.UTF-8

RUN cp -av /usr/bin/pip2 /usr/bin/pip2.7 && \
    pip install setuptools && \
    pip3 install setuptools && \
    pip install wheel && \
    pip3 install wheel

COPY . /home/Osmedeus
WORKDIR /home/Osmedeus

RUN ./install.sh  && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["python3", "server/manage.py", "runserver", "0.0.0.0:8000"]


