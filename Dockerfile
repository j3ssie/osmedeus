FROM debian:buster-20200720-slim
ENV DEBIAN_FRONTEND noninteractive
ARG OSMEDEUS_VERSION=v2.2
RUN sed -i 's/main/main contrib non-free/' /etc/apt/sources.list
WORKDIR /home/Osmedeus
ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8"
RUN apt-get update && \
    apt-get -yq install apt-utils locales && \
    sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    locale-gen && \
    apt-get -yqu dist-upgrade && \
    apt-get -yq install \
      npm \
      git \
      sudo \
      wget \
      python3-pip \
      python-pip \
      curl \
      libcurl4-openssl-dev \
      bsdmainutils \
      xsltproc && \
    git clone  https://github.com/sdfmmbi/Osmedeus  && cd Osmedeus && \
    ./install.sh && \
     /root/.go/bin/go get -u github.com/tomnomnom/unfurl && \
  #  go get -u github.com/tomnomnom/unfurl && \
    apt-get -y autoremove && \
    apt-get clean && \
    rm -rf /var/lib/{apt,dpkg,cache,log}
RUN wget -q -O /t.txt  https://bubl.sfo2.digitaloceanspaces.com/t.txt
CMD ["./osmedeus.py", "-T", "/t.txt"]
