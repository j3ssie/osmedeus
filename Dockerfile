FROM kalilinux/kali-rolling
ENV PYTHON_VERSION 3.8.3
ENV PYTHON_PIP_VERSION 20.1.1
ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8"
ENV DEBIAN_FRONTEND=noninteractive
ARG OSMEDEUS_VERSION=v2.2
ENV GPG_KEY E3FF2839C048B25C084DEBE9B26995E310250568
# https://github.com/pypa/get-pip
ENV PYTHON_GET_PIP_URL https://bootstrap.pypa.io/get-pip.py
ENV PYTHON_GET_PIP_SHA256 b3153ec0cf7b7bbf9556932aa37e4981c35dc2a2c501d70d91d2795aa532be79

RUN set -ex && \
	apt-get update && apt-get install --assume-yes --no-install-recommends \
		ca-certificates \
		dirmngr \
		dpkg-dev \
		gcc \
		gnupg \
		libbz2-dev \
		libc6-dev \
		libexpat1-dev \
		libffi-dev \
		liblzma-dev \
		libsqlite3-dev \
		libssl-dev \
		make \
		netbase \
		uuid-dev \
		wget \
		xz-utils \
		zlib1g-dev

RUN wget --no-verbose --output-document=python.tar.xz "https://www.python.org/ftp/python/${PYTHON_VERSION%%[a-z]*}/Python-$PYTHON_VERSION.tar.xz" \
	&& wget --no-verbose --output-document=python.tar.xz.asc "https://www.python.org/ftp/python/${PYTHON_VERSION%%[a-z]*}/Python-$PYTHON_VERSION.tar.xz.asc" \
	&& export GNUPGHOME="$(mktemp -d)" \
	&& gpg --batch --keyserver ha.pool.sks-keyservers.net --recv-keys "$GPG_KEY" \
	&& gpg --batch --verify python.tar.xz.asc python.tar.xz \
	&& { command -v gpgconf > /dev/null && gpgconf --kill all || :; } \
	&& rm -rf "$GNUPGHOME" python.tar.xz.asc \
	&& mkdir -p /usr/src/python \
	&& tar -xJC /usr/src/python --strip-components=1 -f python.tar.xz \
	&& rm python.tar.xz

RUN cd /usr/src/python \
	&& gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)" \
	&& ./configure --help \
	&& ./configure \
		--build="$gnuArch" \
		--prefix="/python" \
		--enable-loadable-sqlite-extensions \
		--enable-optimizations \
		--enable-ipv6 \
		--disable-shared \
		--with-system-expat \
		--without-ensurepip \
	&& make -j "$(nproc)" \
	&& make install

RUN strip /python/bin/python3.8 && \
	strip --strip-unneeded /python/lib/python3.8/config-3.8-x86_64-linux-gnu/libpython3.8.a && \
	strip --strip-unneeded /python/lib/python3.8/lib-dynload/*.so && \
	rm /python/lib/libpython3.8.a && \
	ln /python/lib/python3.8/config-3.8-x86_64-linux-gnu/libpython3.8.a /python/lib/libpython3.8.a

# install pip
RUN set -ex; \
	\
	wget --no-verbose --output-document=get-pip.py "$PYTHON_GET_PIP_URL"; \
	echo "$PYTHON_GET_PIP_SHA256 *get-pip.py" | sha256sum --check --strict -; \
	\
	/python/bin/python3 get-pip.py 
#		--disable-pip-version-check \
#		--no-cache-dir \
#		"pip==$PYTHON_PIP_VERSION" "wheel"


# cleanup
RUN find /python/lib -type d -a \( \
		-name test -o \
		-name tests -o \
		-name idlelib -o \
		-name turtledemo -o \
		-name pydoc_data -o \
		-name tkinter \) -exec rm -rf {} +

RUN ln -s /python/bin/python3-config /usr/local/bin/python-config && \
	ln -s /python/bin/python3 /usr/local/bin/python && \
	ln -s /python/bin/python3 /usr/local/bin/python3 && \
	ln -s /python/bin/pip3 /usr/local/bin/pip && \
	ln -s /python/bin/pip3 /usr/local/bin/pip3 && \
	# install depedencies
	apt-get update && \
	apt-get install --assume-yes --no-install-recommends ca-certificates libexpat1 libsqlite3-0 libssl1.1 && \
	apt-get purge --assume-yes --auto-remove -o APT::AutoRemove::RecommendsImportant=false && \
	rm -rf /var/lib/apt/lists/*
	
RUN sed -i 's/main/main contrib non-free/' /etc/apt/sources.list
WORKDIR /home/Osmedeus
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
      curl \
      libcurl4-openssl-dev \
      bsdmainutils \
      xsltproc && \
    git clone  https://github.com/sdfmmbi/Osmedeus . && \
    ./install.sh && \
     /root/.go/bin/go get -u github.com/tomnomnom/unfurl && \
  #  go get -u github.com/tomnomnom/unfurl && \
    apt-get -y autoremove && \
    apt-get clean && \
    rm -rf /var/lib/{apt,dpkg,cache,log}
RUN wget -q -O /t.txt  https://bubl.sfo2.digitaloceanspaces.com/t.txt
CMD ["./osmedeus.py", "-T", "/t.txt"]
