# Docker image for gomobile
#
# http://godoc.org/golang.org/x/mobile/cmd/gomobile

FROM debian:jessie
MAINTAINER Mathias Monnerville <matm@outofpluto.com>

# ia32 libs required for Android tools
RUN dpkg --add-architecture i386
RUN apt-get -qq update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -q -y \
    vim \
    git \
    gcc \
    openjdk-7-jdk \
    ant \
    make \
    wget \
    libncurses5:i386 \
    libstdc++6:i386 \
    zlib1g:i386

# go user
RUN useradd -d /home/go -m -s /bin/bash go

# Build 1.4.2 first
WORKDIR /usr/local/go
RUN git clone https://go.googlesource.com/go 14
# Prepare source tree for Go 1.5
RUN cp -r 14 15
WORKDIR /usr/local/go/14
RUN git checkout go1.4.2
WORKDIR /usr/local/go/14/src
RUN ./make.bash

# Build Go from tip
ENV GOROOT_BOOTSTRAP /usr/local/go/14
WORKDIR /usr/local/go/15/src
RUN ./make.bash

ENV GOROOT /usr/local/go/15
ENV PATH /home/go/bin:/usr/local/go/15/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Install Android SDK
WORKDIR /usr/local
RUN wget http://dl.google.com/android/android-sdk_r24.0.2-linux.tgz
RUN tar xzf android-sdk_r24.0.2-linux.tgz
RUN chown -R root:staff /usr/local/android-sdk-linux/
# Required distribution for building the Android application (android list sdk --all)
# 1: Android SDK Tools, revision 24.0.2
# 2: Android SDK Platform-tools, revision 21
# 3: Android SDK Build-tools, revision 21.1.2
# 25: SDK Platform Android 4.1.2, API 16, revision 5
# 85: Google APIs, Android API 16, revision 3
# 113: Android Support Library, revision 21.0.3
RUN echo yes | /usr/local/android-sdk-linux/tools/android update sdk -u -a -t 1,2,3,25,85,113

# Install Android NDK
RUN wget http://dl.google.com/android/ndk/android-ndk-r10d-linux-x86_64.bin
RUN chmod a+x android-ndk-r10d-linux-x86_64.bin
RUN ./android-ndk-r10d-linux-x86_64.bin
RUN rm android-ndk-r10d-linux-x86_64.bin
RUN rm -f *.tgz *.tar.gz *.bin

# gomobile init
ENV ANDROID_HOME /usr/local/android-sdk-linux
ENV GOPATH /home/go

USER go
WORKDIR /home/go
RUN mkdir src bin pkg
RUN go get golang.org/x/mobile/cmd/gomobile

# Init as root, some files need to be installed in /usr/local
USER root
RUN /home/go/bin/gomobile init
RUN chown -R go:go /home/go

USER go
ENTRYPOINT ["gomobile"]
