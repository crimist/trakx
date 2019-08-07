#!/usr/bin/env bash

# Update
git pull

# Setup root if not setup
mkdir -p ~/.trakx/
cp -n config.yaml ~/.trakx/config.yaml
cp -n index.html ~/.trakx/index.html

# install
go install -gcflags='-l=4 -s' -tags expvar
