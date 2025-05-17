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

free -h

cd /app
go build -o "$tmpFile" ./cmd/myshell && echo "done" || echo "failed"

echo "running"

"$tmpFile" "$@"

echo "done"
