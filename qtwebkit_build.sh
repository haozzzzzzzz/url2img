#!/bin/bash
mkdir -p WebkitBuild/Release
cd WebkitBuild/Release

cmake -DPORT=Qt -DCMAKE_BUILD_TYPE=Release \
    -DSKIP_DWZ=ON \
    -DCMAKE_INSTALL_PREFIX=/data/Qt/5.15.0/gcc_64/ \
    -DQt_DIR=/data/Qt/5.15.0/gcc_64/lib/cmake/Qt5 ../..

make -j4 # 4 cpu core
make install

