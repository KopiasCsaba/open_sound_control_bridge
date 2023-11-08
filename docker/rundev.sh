#!/bin/bash
# DOC: This script starts the app, and also restarts it upon any file change in src, enabling fast and efficient development.
# The app is also restarted upon changes in the "docker/devenv/" folder.

# Jump to the src directory, with a relative cd to this script's location.
cd "$(dirname "$0")/../src" || exit

export PATH="${PATH}:${HOME}/go/bin"
git config --global --add safe.directory /mnt

echo "> Downloading dependencies ..."
go mod download



export GOOS=linux
export GOARCH=amd64

while true; do
  echo "> Running gofumt..."
  gofumpt -w . &

  echo "> Running linters..."
  golangci-lint run --color always &

  echo "> Starting app..."
  # shellcheck disable=SC2086
  go run -ldflags "-s -w -X main.Revision='devel' -X main.BuildTime=$(date +'%Y-%m-%d_%T')" cmd/main/main.go &

  # Store PID
  PID=$!


  echo "Started app with PID: $PID. Restarting upon any file change."
  # Wait on app to start really before checking for changes
  go build  -o ../build/oscbridge -tags netgo -ldflags "-s -w -X main.Revision='$(git rev-parse HEAD 2>/dev/null || echo 1)' -X main.BuildTime=$(date +'%Y-%m-%d_%T')" cmd/main/main.go  &
  sleep 2
  # Wait on file changes
  inotifywait -e modify -e move -e create -e delete -e attrib -r "$(pwd)" "$(pwd)/../docker/devenv/" >/dev/null 2>&1

  echo "Restarting..."
  # Kill the app, repeat...
  kill -9 "$PID" > /dev/null 2>&1
  pkill -f go-build  > /dev/null  2>&1 # For some reason descendant processes would stay alive

done
