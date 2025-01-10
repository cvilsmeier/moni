#!/bin/sh

set -e
stat go.mod > /dev/null # must be in src dir

rm -rf _dist
mkdir -p _dist/linux _dist/windows

echo "go build"
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o _dist/linux .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o _dist/windows .

echo "tar"
tar -czf _dist/moni-linux.tar.gz   README.md -C _dist/linux   moni
tar -czf _dist/moni-windows.tar.gz README.md -C _dist/windows moni.exe

