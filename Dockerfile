FROM debian:buster-20191014-slim
ARG OSMEDEUS_VERSION=v.2.1
RUN sed -i 's/main/main contrib non-free/' /etc/apt/sources.list
WORKDIR /home/Osmedeus
ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8"
RUN apt-get update && \
    apt-get -yu dist-upgrade && \
    apt-get -qq install npm && \
    apt-get -qq install locales git sudo wget python3-pip python-pip curl libcurl4-openssl-dev bsdmainutils xsltproc && \
    sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    locale-gen && \
    cp -av /usr/bin/pip2 /usr/bin/pip2.7 && \
    git clone --depth 1 https://github.com/sdfmmbi/Osmedeus -b $OSMEDEUS_VERSION . && \
    ./install.sh && \
    go get -u github.com/tomnomnom/unfurl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
EXPOSE 8000
CMD ["./osmedeus.py" -T data/t.txt"]
