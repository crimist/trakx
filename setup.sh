#!/usr/bin/env bash

# Update
env GIT_TERMINAL_PROMPT=1 go get -u -v github.com/syc0x00/trakx

# Setup root if not setup
mkdir -p ~/.trakx/
cp -n config.yaml ~/.trakx/config.yaml
cp -n index.html ~/.trakx/index.html
