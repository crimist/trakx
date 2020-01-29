#!/usr/bin/env bash
root=/usr/local/etc/trakx
config=trakx.yaml
index=trakx.html

git pull
mkdir -p $root
cp -n $config $root/$config
cp -n $index $root/$index
go install -v -gcflags='-l=4' ../

updateConf() {
    cp $config $root/$config
    echo "Updated config file"
}

updateIndex() {
    cp $index $root/$index
    echo "Updated index file"
}

if ! cmp $config $root/$config >/dev/null 2>&1; then
    read -p "Config file differs, overwrite? (y/n): " -n 1 -r; echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        updateConf
    fi
else
    updateConf
fi

if ! cmp $index $root/$index >/dev/null 2>&1; then
    read -p "Index file differs, overwrite? (y/n): " -n 1 -r; echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        updateIndex
    fi
else
    updateIndex
fi
