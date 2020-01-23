#!/usr/bin/env bash
root=/usr/local/etc/trakx
config=trakx.yaml
index=trakx.html

git pull
mkdir -p $root
cp -n $config $root/$config
cp -n $index $root/$index
go install -v -gcflags='-l=4' ../

if ! cmp $config $root/$config >/dev/null 2>&1; then
    read -p "Config file differs, overwrite? (y/n): " -n 1 -r; echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        cp $config $root/$config
    fi
else
    cp $config $root/$config
fi

if ! cmp $index $root/$index >/dev/null 2>&1; then
    read -p "Index file differs, overwrite? (y/n): " -n 1 -r; echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        cp $index $root/$index
    fi
else
    cp $index $root/$index
fi
