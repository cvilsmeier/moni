#!/bin/sh

set -e
stat go.mod > /dev/null   # must be in src dir

git status
go test -v -count 1 --tags=check ./... 
staticcheck ./...
echo "check ok"
