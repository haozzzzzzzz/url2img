#!/bin/bash
cd /home/hao/Project/qt/qtwebkit/qtwebkit-5.212.0-alpha4
mkdir -p WebkitBuild/Release
cd WebkitBuild/Release

echo 'export CMAKE_INSTALL_PREFIX=/home/hao/Qt/5.15.0/gcc_64/' >> ~/.bashrc
source ~/.bashrc

cmake -DPORT=Qt -DCMAKE_INSTALL_PREFIX=/home/hao/Qt/5.15.0/gcc_64 -DCMAKE_BUILD_TYPE=Release \
	-DQt5_DIR=/home/hao/Qt/5.15.0/gcc_64/lib/cmake/Qt5 ../..

make -j4
make install
