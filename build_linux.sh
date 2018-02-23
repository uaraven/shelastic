#!/bin/bash

echo "Building shelastic for Linux amd64"
env GOOS=linux GOARCH=amd64 go build
