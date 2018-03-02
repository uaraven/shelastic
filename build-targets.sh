#!/bin/bash

ARCH=amd64

echo "Building shelastic for Linux amd64"
mkdir -p bin/linux-$ARCH
env GOOS=linux GOARCH=$ARCH go build -o bin/linux-$ARCH/shelastic

echo "Building shelastic for MacOS amd64"
mkdir -p bin/macos-$ARCH
env GOOS=darwin GOARCH=$ARCH go build -o bin/macos-$ARCH/shelastic

echo "Building shelastic for Windows amd64"
mkdir -p bin/win-$ARCH
env GOOS=windows GOARCH=$ARCH go build -o bin/win-$ARCH/shelastic.exe