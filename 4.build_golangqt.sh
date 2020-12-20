#!/bin/bash

# https://github.com/therecipe/qt/wiki/Installation-on-Linux
# 进去~/.bashrc添加环境变量
# export QT_DIR=/home/hao/Qt
# export QT_VERSION=5.15.0
# export QT_API=5.13.0
# export QT_WEBKIT=true

sudo apt-get -y install build-essential libglu1-mesa-dev libpulse-dev libglib2.0-dev

go get -v -tags=no_env github.com/therecipe/qt/cmd/...
$(go env GOPATH)/bin/qtsetup


