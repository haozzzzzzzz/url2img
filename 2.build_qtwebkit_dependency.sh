#!/bin/bash
# https://packages.debian.org/source/sid/qtwebkit-opensource-src
sudo apt intsall -y cmake \
	ninja-build \
	libsqlite3-dev \
	libicu-dev \
	ruby \
	perl \
	python2.7 \
	bison \
	flex \
	gperf \
	libxml2 \
	libxslt-dev \
	libjepg-dev \
	libpng-dev \
	zlib1g-dev \
	libhyphen-dev \
	libglib2.0-dev \
	libxcomposite-dev \
	libxrender-dev \
	libwoff-dev \
	pkg-kde-tools \
	xvfb \
	libdouble-conversion-dev

sudo apt-get install libgstreamer1.0-0 gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-libav gstreamer1.0-doc gstreamer1.0-tools gstreamer1.0-x gstreamer1.0-alsa gstreamer1.0-gl gstreamer1.0-gtk3 gstreamer1.0-qt5 gstreamer1.0-pulseaudio libgstreamer-plugins-base1.0-dev

sudo apt-get install -y fontconfig-config
sudo apt install -y libwebp-de
sudo apt install -y mesa-common-dev libglu1-mesa-dev
