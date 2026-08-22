#!/bin/sh
# Use the locally-installed Go (not on system PATH).
export GOROOT=/c/GoInstall/go
export GOFLAGS=-mod=mod
export GOSUMDB=off
export PATH="$GOROOT/bin:$PATH"
exec go "$@"
