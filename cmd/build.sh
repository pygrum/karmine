#!/bin/bash
set -e

# Use this script to rebuild the commands if any changes were made post-installation

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

cd $SCRIPTPATH

if [ ! -d "../../kbin" ]; then
    mkdir "../../kbin"
fi

for i in $(ls); do
    if [ ! -f $i ]; then
        cd $i && 
        go build && 
        mv $i ../../kbin && cd ..;
    fi
done

if [ -d "${HOME}/.kbin" ]; then
    rm -rf "${HOME}/.kbin"
fi

mv "../../kbin" ${HOME}/.kbin