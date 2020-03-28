#!/bin/bash
set -e

: "${ZSM_TESTCOVERAGE_FILE:=$PWD/.coverage.out}"
: "${ZSM_TESTCOVERAGE_TESTFLAGS:=-race}"

if [ $EUID = 0 ]; then
    echo "You can't be root"
    exit 1
fi

if [ ! -d "$PWD/.git" ] || [ ! -f "$PWD/$0" ]; then
    echo "$0 needs to be called from the project root."
    exit 1
fi

if ! command -v go > /dev/null 2>&1; then
    echo "No go executable found"
    exit 1
fi

GO_PACKAGES=$(command go list ./... | grep -v scripts | tr "\n" ",")
command go test \
    "$ZSM_TESTCOVERAGE_TESTFLAGS" \
    -p 1 \
    -count=1 \
    -covermode=atomic \
    -coverprofile="$ZSM_TESTCOVERAGE_FILE" \
    -coverpkg="$GO_PACKAGES" \
    ./... 2> /dev/null

case $1 in
    html)
        command go tool cover -html="$ZSM_TESTCOVERAGE_FILE"
        ;;
    *)
        command go tool cover -func="$ZSM_TESTCOVERAGE_FILE" | tail -n1
        ;;
esac
