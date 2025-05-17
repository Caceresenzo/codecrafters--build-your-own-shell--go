#!/bin/sh
#
# DON'T EDIT THIS!
#
# CodeCrafters uses this file to test your code. Don't make any changes here!
#
# DON'T EDIT THIS!
set -e
# set -x

tmpFile=$(mktemp)

# echo "cd $(dirname "$0") && go build -o "$tmpFile" ./cmd/myshell"
( cd $(dirname "$0") &&
	go build -o "$tmpFile" ./cmd/myshell >/dev/stderr)

"$tmpFile" "$@"
