#!/bin/bash
set -e

CONFIG_LINES=(
    "{"
    "\"bin_path\":\"${HOME}/.kbin\","
    "\"src_path\":\"$(realpath .)\","
    "\"sql_user\":\"karmine\","
    "\"sql_pass\":\"y/k9xbw0Jb61Q9/r\","
    "\"cert_pem\":\"${HOME}/.kdots/karmine.crt\","
    "\"key_pem\":\"${HOME}/.kdots/karmine.key\","
    "\"endpoint\":\"/api\""
    "}"
)

go build 

if [ -f "${HOME}/.konfig" ]
then
    echo ".konfig found at ${HOME}. skipping"
    exit 1
fi

for LINE in ${CONFIG_LINES[@]}
do
    echo "$LINE" >> ${HOME}/.konfig
done

if [ -d "${HOME}/.kbin" ]
then
    echo ".kbin already present. exiting"
    exit 1
fi

if [ ! -d "${HOME}/.kdots" ]
then
    mkdir ${HOME}/.kdots
    sudo openssl req -x509 -nodes -days 365 \
                -subj "/CN=$n" \
                  -addext "subjectAltName = DNS:$n" \
                  -newkey rsa:2048 -keyout ${HOME}/.kdots/karmine.key -out ${HOME}/.kdots/karmine.crt
    # sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${HOME}/.kdots/karmine.key -out ${HOME}/.kdots/karmine.crt
fi

if [ -d "${HOME}/.kbin" ]
then
    rm -rf ${HOME}/.kbin
fi
cp -r kbin ${HOME}/.kbin

echo "Complete. Go to https://github.com/pygrum/karmine for full setup instructions"