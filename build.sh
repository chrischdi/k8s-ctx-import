#!/usr/bin/env bash

set -e

platforms="darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64 windows/386 windows/amd64"

name="k8s-ctx-import"

mkdir -p bin tar

for platform in ${platforms}
do
  split=(${platform//\// })
  goos=${split[0]}
  goarch=${split[1]}

  # prepare
  ext=""
  if [ "$goos" == "windows" ]; then
    ext=".exe"
  fi
  mkdir -p bin/$goos/$goarch

  # build
  CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags='-s -w' -v -o bin/$goos/$goarch/$name$ext
  
  # pack
  tar cfvz tar/$name-$goos-$goarch.tar.gz -C bin/$goos/$goarch .
done