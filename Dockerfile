# Dockerfile to build an image with the local version of go.mobile.
#
#  > docker build -t mobile $GOPATH/src/golang.org/x/mobile
#  > docker run -it --rm -v $GOPATH/src:/src mobile

FROM ubuntu:12.04

# Install system-level dependencies.
ENV DEBIAN_FRONTEND noninteractive
RUN echo "debconf shared/accepted-oracle-license-v1-1 select true" | debconf-set-selections && \
	echo "debconf shared/accepted-oracle-license-v1-1 seen true" | debconf-set-selections
RUN apt-get update && \
	apt-get -y install build-essential python-software-properties bzip2 unzip curl \
		git subversion mercurial bzr \
		libncurses5:i386 libstdc++6:i386 zlib1g:i386 && \
	add-apt-repository ppa:webupd8team/java && \
	apt-get update && \
	apt-get -y install oracle-java6-installer

# Install Ant.
RUN curl -L http://archive.apache.org/dist/ant/binaries/apache-ant-1.9.2-bin.tar.gz | tar xz -C /usr/local
ENV ANT_HOME /usr/local/apache-ant-1.9.2

# Install Android SDK.
RUN curl -L http://dl.google.com/android/android-sdk_r23.0.2-linux.tgz | tar xz -C /usr/local
ENV ANDROID_HOME /usr/local/android-sdk-linux
RUN echo y | $ANDROID_HOME/tools/android update sdk --no-ui --all --filter build-tools-19.1.0 && \
	echo y | $ANDROID_HOME/tools/android update sdk --no-ui --all --filter platform-tools && \
	echo y | $ANDROID_HOME/tools/android update sdk --no-ui --all --filter android-19

# Install Android NDK.
RUN curl -L http://dl.google.com/android/ndk/android-ndk-r9d-linux-x86_64.tar.bz2 | tar xj -C /usr/local
ENV NDK_ROOT /usr/local/android-ndk-r9d
RUN $NDK_ROOT/build/tools/make-standalone-toolchain.sh --platform=android-9 --install-dir=$NDK_ROOT --system=linux-x86_64

# Install Gradle 2.1
# : android-gradle compatibility
#   http://tools.android.com/tech-docs/new-build-system/version-compatibility
RUN curl -L http://services.gradle.org/distributions/gradle-2.1-all.zip -o /tmp/gradle-2.1-all.zip && unzip /tmp/gradle-2.1-all.zip -d /usr/local && rm /tmp/gradle-2.1-all.zip
ENV GRADLE_HOME /usr/local/gradle-2.1

# Update PATH for the above.
ENV PATH $PATH:$ANDROID_HOME/tools
ENV PATH $PATH:$ANDROID_HOME/platform-tools
ENV PATH $PATH:$NDK_ROOT
ENV PATH $PATH:$ANT_HOME/bin
ENV PATH $PATH:$GRADLE_HOME/bin

# Install Go.
ENV GOROOT /go
ENV GOPATH /
ENV PATH $PATH:$GOROOT/bin
RUN curl -L https://github.com/golang/go/archive/master.zip -o /tmp/go.zip && \
	unzip /tmp/go.zip && \
	rm /tmp/go.zip && \
	mv /go-master $GOROOT && \
	echo devel > $GOROOT/VERSION && \
	cd $GOROOT/src && \
	./all.bash && \
	CC_FOR_TARGET=$NDK_ROOT/bin/arm-linux-androideabi-gcc GOOS=android GOARCH=arm GOARM=7 ./make.bash

WORKDIR $GOPATH/src/golang.org/x/mobile
