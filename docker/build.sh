#!/bin/bash
# DOC: This script builds the app itself, this is executed inside of the container

# Jump to the src directory, with a relative cd to this script's location.
cd "$(dirname "$0")/../src" || exit

BUILD_DIR="/mnt/build/"
export GOPATH=""

echo -e "Downloading dependencies...\n"
go mod download

# build_application($1,$2,$3) executes go build with all the required parameters.
# $1: The selected go operating system
# $2: The selected go architecture
build_application() {
  #  set -x
  export CGO_ENABLED=0
  export GOOS=$1
  export GOARCH=$2
  export EXTENSION=$3

  echo "BUILDING for $GOOS/$GOARCH"

  rm "$BUILD_DIR/app-$GOOS-$GOARCH.bin" >/dev/null 2>&1

  go build -tags netgo \
    -ldflags "-s -w -X main.Revision=$(git rev-parse HEAD 2>/dev/null || echo 1) -X main.BuildTime=$(date +'%Y-%m-%d_%T') " \
    -o "$BUILD_DIR/app-$GOOS-$GOARCH.$EXTENSION" \
    cmd/main/main.go


  echo -e "\n"
}
# BUILD
# ------------------------------------------

# Normal compilation for linux/amd64 platform
export CC=gcc
build_application linux amd64 bin
build_application windows amd64 exe

echo "BUILT ARTIFACTS:"
ls -lh "$BUILD_DIR"
