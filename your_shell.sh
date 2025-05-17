#!/bin/sh
#
# DON'T EDIT THIS!
#
# CodeCrafters uses this file to test your code. Don't make any changes here!
#
# DON'T EDIT THIS!
set -e
set -x

tmpFile=$(mktemp)

echo "building"

sleep 1
echo "building2"

sleep 3
echo "building3"


cd /app
go build -o "$tmpFile" ./cmd/myshell

echo "running"

"$tmpFile" "$@"

echo "done"
