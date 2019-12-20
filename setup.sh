#!/usr/bin/env bash

# Update
git pull

# Setup root if not setup
mkdir -p /usr/local/trakx/
cp -n trakx.yaml /usr/local/trakx/trakx.yaml
cp index.html /usr/local/trakx/index.html

# install
go install -gcflags='-l=4 -s'
