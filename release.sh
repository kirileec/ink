#!/bin/sh

# build binary releases for ink

build () {
  echo "building for $1 $2..."
  suffix=""
  if [ $1 == "windows" ]
  then
    suffix=".exe"
  fi
  GOOS=$1 GOARCH=$2 go build -o release/ink$suffix
  cd ..
}

rm -rf release
mkdir -p release

build linux amd64
build darwin amd64
build windows amd64

# rm -rf release/blog
