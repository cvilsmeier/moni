#!/bin/sh

set -e
stat go.mod > /dev/null   # must be in src dir

rm -rf _dist
