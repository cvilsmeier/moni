#!/bin/sh

set -e
stat go.mod > /dev/null   # must be in src dir

git status
go test ./... -count 1
staticcheck ./...
echo "check ok"
