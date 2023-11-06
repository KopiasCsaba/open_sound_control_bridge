#!/bin/bash
# This script performs LINT checks on the source.

export PATH="${PATH}:${HOME}/go/bin"

# Jump to the src directory, with a relative cd to this script's location.
cd "$(dirname "$0")/../src" || exit


echo "> Downloading dependencies ..."
go mod download

echo "> Running linters..."
golangci-lint run --color always